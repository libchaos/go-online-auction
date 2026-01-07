import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-blue-600 mb-4">Auction Frontend Demo</h1>
        <p className="text-gray-600 mb-4">TailwindCSS is working!</p>
        <button 
          onClick={() => setCount((count) => count + 1)}
          className="bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 px-4 rounded"
        >
          Count is {count}
        </button>
        <p className="mt-4 text-sm text-gray-500">
          API Base URL: {import.meta.env.VITE_API_BASE_URL}
        </p>
      </div>
    </div>
  )
}

export default App

