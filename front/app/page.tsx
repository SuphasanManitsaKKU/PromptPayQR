"use client";

import React, { useState } from "react";
import axios from "axios";
import Swal from "sweetalert2";
import { useRouter } from "next/navigation";

export default function Home() {
  const [amount, setAmount] = useState<string>("");
  const [qrCode, setQrCode] = useState<string>("");
  const [timestamp, setTimestamp] = useState<string>("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const router = useRouter(); // ใช้ useRouter สำหรับการนำทาง

  const generateQRCode = async (event: any) => {
    event.preventDefault();
    const parsedAmount = parseFloat(amount);

    // Validation: Ensure input is a number and greater than 0
    if (!amount || isNaN(parsedAmount) || parsedAmount <= 0) {
      Swal.fire("Error", "กรุณากรอกจำนวนเงินที่ถูกต้อง (มากกว่า 0)", "error");
      return;
    }

    try {
      const response = await axios.post(`${process.env.NEXT_PUBLIC_API_URL}/generateQR`, {
        amount: parsedAmount,
      });

      // console.log(response.data);

      const { RespCode, Result, Timestamp } = response.data;
      if (RespCode === 200) {
        setQrCode(Result);
        setTimestamp(Timestamp);
      } else {
        Swal.fire("Error", "เกิดข้อผิดพลาดในการสร้าง QR Code", "error");
      }
    } catch (error) {
      console.error("Error:", error);
      Swal.fire("Error", "ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้", "error");
    }
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedFile(event.target.files[0]);
    }
  };

  const checkSlip = async (event: any) => {
    event.preventDefault();

    const formData = new FormData();
    const fileInput = document.querySelector("input[type='file']") as HTMLInputElement;
    if (!fileInput || !fileInput.files || !fileInput.files[0]) {
      Swal.fire("Error", "กรุณาแนบไฟล์รูปภาพ", "error");
      return;
    }

    formData.append("file", fileInput.files[0]);
    formData.append("amount", amount);
    formData.append("timestamp", timestamp);

    const response = await axios.post(`${process.env.NEXT_PUBLIC_API_URL}/uploadSlip`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });

    // console.log(response.data);

    const { RespCode, RespMessage } = response.data;
    if (RespCode === 200) {
      Swal.fire("สำเร็จ", "สลิปได้รับการตรวจสอบเรียบร้อย", "success").then(() => {
        router.push("/thankYou"); // นำทางไปหน้า thankYou
      });
    } else {
      Swal.fire("Error", RespMessage);
    }
  };

  const saveQRCode = () => {
    const link = document.createElement("a");
    link.href = qrCode; // Use the generated QR code URL
    link.download = "promptpay-qr-code.png"; // Set a default filename
    link.click(); // Trigger the download
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50">
      <div className="bg-white p-6 rounded-lg shadow-lg w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-4">
          สร้าง PromptPay QR Code
        </h1>
        <form onSubmit={generateQRCode}>
          <input
            type="text"
            placeholder="กรอกจำนวนเงิน"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full p-3 border rounded-lg mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            className="w-full bg-blue-500 text-white py-3 rounded-lg font-semibold hover:bg-blue-600"
          >
            สร้าง QR Code
          </button>
        </form>
        {qrCode && (
          <div className="mt-6 text-center">
            <img
              src={qrCode}
              alt="QR Code"
              className="w-64 h-64 mx-auto rounded-lg border shadow-md"
            />
            <div className="mt-4">
              <button
                onClick={saveQRCode}
                className="w-full bg-yellow-500 text-white py-3 rounded-lg font-semibold hover:bg-yellow-600"
              >
                บันทึก QR Code
              </button>
            </div>

            <form onSubmit={checkSlip} className="mt-4">
              <label className="block mb-2 text-sm font-medium text-gray-700">
                อัปโหลด Slip
              </label>
              <input
                type="file"
                accept="image/*"
                onChange={handleFileChange}
                className="w-full p-2 border rounded-lg mb-4"
              />
              <button
                type="submit"
                className="w-full bg-green-500 text-white py-3 rounded-lg font-semibold hover:bg-green-600"
              >
                ส่ง Slip
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
