package main

import (
	"PromptPayQR/model"
	"PromptPayQR/repository"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	promptpayqr "github.com/kazekim/promptpay-qr-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// https://youtu.be/psBmEOIGF6c?si=EUv2j9pQ5S-QE7EJ

func main() {
	// DSN for connecting to MySQL]
	dsn_localhost := "root:1234@tcp(localhost:3306)/gorm_db?charset=utf8mb4&parseTime=True&loc=Local"
	dsn_mysql := "root:1234@tcp(mysql:3306)/gorm_db?charset=utf8mb4&parseTime=True&loc=Local"
	var db *gorm.DB
	var err error

	// Try connecting 30 times
	for i := 0; i < 30; i++ {
		if i%2 == 0 {
			db, err = gorm.Open(mysql.Open(dsn_localhost), &gorm.Config{})
			if err == nil {
				fmt.Println("Connected to the database successfully!")
				break
			}
		} else {
			db, err = gorm.Open(mysql.Open(dsn_mysql), &gorm.Config{})
			if err == nil {
				fmt.Println("Connected to the database successfully!")
				break
			}
		}
		// Log the error and wait for 1 second before retrying
		log.Printf("Attempt %d: Failed to connect to the database, retrying...\n", i+1)
		time.Sleep(1 * time.Second)
	}

	// If after 30 retries the connection is still unsuccessful, return an error
	if err != nil {
		log.Fatal("Unable to connect to the database after 30 attempts:", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&model.Slip{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	App := fiber.New()
	App.Use(cors.New(cors.ConfigDefault))
	app := App.Group("/api")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"RespCode":    200,
			"RespMessage": "Welcome to PromptPayQR API",
		})
	})

	app.Post("/generateQR", func(c *fiber.Ctx) error {
		type Request struct {
			Amount float64 `json:"amount"`
		}

		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RespCode":    400,
				"RespMessage": "Invalid request body",
			})
		}

		mobileNumber := "0934178026"
		amount := fmt.Sprintf("%.2f", req.Amount)

		qrCodeBytes, err := promptpayqr.QRForTargetWithAmount(mobileNumber, amount)
		if err != nil {
			log.Println("Error generating QR code:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"RespCode":    500,
				"RespMessage": "Failed to generate QR code",
			})
		}

		// Dereference pointer to get the actual byte slice
		base64Image := byteToBase64(*qrCodeBytes)

		timestamp := fmt.Sprintf("%s", time.Now().Format(time.RFC3339))

		return c.JSON(fiber.Map{
			"RespCode":    200,
			"RespMessage": "QR Code generated successfully",
			"Result":      "data:image/png;base64," + base64Image,
			"Timestamp":   timestamp,
			"Amount":      amount,
		})

	})

	app.Post("/uploadSlip", func(c *fiber.Ctx) error {
		// ดึงไฟล์จาก request
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RespCode":    400,
				"RespMessage": "กรุณาแนบไฟล์รูปภาพ",
			})
		}

		amount := c.FormValue("amount")
		if amount == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RespCode":    400,
				"RespMessage": "กรุณากรอกจำนวนเงิน",
			})
		}
		timestamp := c.FormValue("timestamp")
		if timestamp == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"RespCode":    400,
				"RespMessage": "กรุณากรอกเวลาที่โอนเงิน",
			})
		}

		// เปิดไฟล์เพื่อเตรียมส่งไปยัง EasySlip API
		fileContent, err := file.Open()
		if err != nil {
			return c.JSON(fiber.Map{
				"RespCode":    500,
				"RespMessage": "ไม่สามารถเปิดไฟล์ได้",
			})
		}
		defer fileContent.Close()

		// ส่งไฟล์ไปยัง EasySlip API
		resp, err := verifySlip(db, fileContent, file.Filename, amount, timestamp)
		if err != nil {
			log.Println("RespMessage:", err)
			verifySlipError := fmt.Sprintf("%v", err)
			return c.JSON(fiber.Map{
				"RespCode":    500,
				"RespMessage": verifySlipError,
			})
		}

		// ตอบกลับผลลัพธ์จาก EasySlip API
		return c.JSON(fiber.Map{
			"RespCode":    200,
			"RespMessage": "ตรวจสอบสลิปสำเร็จ",
			"Result":      resp,
			"Timestamp":   timestamp,
		})
	})

	log.Fatal(App.Listen(":8000"))
}

func byteToBase64(imgByte []byte) string {
	buffer := new(bytes.Buffer)
	buffer.Write(imgByte) // ใช้โดยตรง ไม่ต้องแปลงเป็น Image
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

// ฟังก์ชันสำหรับส่งรูปภาพไปยัง EasySlip API และตรวจสอบข้อมูล
func verifySlip(db *gorm.DB, file io.Reader, filename, amount, timestamp string) (map[string]interface{}, error) {
	// เตรียม URL และ Authorization
	url := "https://developer.easyslip.com/api/v1/verify"
	authToken := "d823a3a2-7d6f-482b-84bd-8d7dfaa80c5b" // หมดอายุแล้วงับฟุ้ววววววว

	// สร้าง request สำหรับ multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่าไฟล์ที่ส่งมามีข้อมูลหรือไม่
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("ไม่สามารถคัดลอกข้อมูลไฟล์ได้: %v", err)
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("ไฟล์ว่างเปล่า")
	}

	// เขียนข้อมูลลงใน form
	_, err = io.Copy(part, buf)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถคัดลอกข้อมูลไปยัง form ได้: %v", err)
	}

	// ปิด writer
	writer.Close()

	// สร้าง HTTP request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// ส่ง request และรับ response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// อ่านและแปลง response เป็น JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// ตรวจสอบค่าที่ได้จาก API
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("การยืนยันการชำระเงินผ่านสลิปไม่สามารถใช้งานได้ในขณะนี้เนื่องจากโควตาของ API ของ EasySlip หมดลงแล้ว")
	}

	transRef := data["transRef"].(string)
	if transRef == "" {
		return nil, fmt.Errorf("transRef ไม่มีค่า")
	}

	// 1. ตรวจสอบว่า transRef ไม่ซ้ำ
	existingSlip, err := repository.GetSlipByTransRef(db, transRef)
	if err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบ transRef: %v", err)
	}
	if existingSlip != nil {
		return nil, fmt.Errorf("transRef นี้มีอยู่ในระบบแล้ว")
	}

	// 2. ตรวจสอบชื่อผู้รับ
	receiver := data["receiver"].(map[string]interface{})
	receiverAccount := receiver["account"].(map[string]interface{})
	receiverName := receiverAccount["name"].(map[string]interface{})["en"].(string)
	if receiverName != "SUPHASAN M" {
		return nil, fmt.Errorf("ชื่อผู้รับไม่ตรงกับที่กำหนด")
	}

	// 3. ตรวจสอบจำนวนเงิน
	amountData := data["amount"].(map[string]interface{})
	amountReceived := amountData["amount"].(float64)
	amountSent, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถแปลงจำนวนเงินที่ส่งไปได้: %v", err)
	}
	if amountReceived != amountSent {
		return nil, fmt.Errorf("จำนวนเงินไม่ตรงกัน")
	}

	// 4. ตรวจสอบวันที่
	dateStr := data["date"].(string)
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถแปลงวันที่ได้: %v", err)
	}
	startTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถแปลงเวลาที่รับเข้ามาได้: %v", err)
	}
	endTime := startTime.Add(10 * time.Minute)
	if date.Before(startTime) || date.After(endTime) {
		return nil, fmt.Errorf("วันที่ไม่อยู่ในช่วงเวลาที่กำหนด")
	}

	// 5. บันทึกข้อมูลลงในฐานข้อมูล
	repository.CreateSlip(db, transRef)

	return result, nil
}
