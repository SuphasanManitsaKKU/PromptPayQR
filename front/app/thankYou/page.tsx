import Link from 'next/link'

export default function Home() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-100 to-purple-200">
      <div className="max-w-md p-8 bg-white rounded-lg shadow-lg text-center">
        <h1 className="text-4xl font-bold text-blue-600">ขอบคุณที่สั่งซื้อ!</h1>
        <p className="mt-4 text-gray-600">คำสั่งซื้อของคุณได้รับการยืนยันเรียบร้อยแล้ว เราหวังว่าคุณจะเพลิดเพลินกับสินค้าของเรา</p>
        <div className="mt-6">
          <Link href="/">
            <button
              className="w-full px-4 py-2 text-lg font-medium text-white bg-blue-500 rounded-lg hover:bg-blue-600 transition duration-200"
            >
              ซื้ออีกครั้ง
            </button>
          </Link>
        </div>
      </div>
    </div>
  );
}
