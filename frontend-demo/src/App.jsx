import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import AuctionListPage from './pages/AuctionListPage';
import CreateAuctionPage from './pages/CreateAuctionPage';
import AuctionDetailPage from './pages/AuctionDetailPage';
import WebSocketPage from './pages/WebSocketPage';

/**
 * Layout component with navigation header
 */
function Layout({ children }) {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation Header */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-8">
              <Link to="/" className="text-xl font-bold text-blue-600 hover:text-blue-700">
                Auction Demo
              </Link>
              <div className="hidden sm:flex gap-4">
                <Link
                  to="/"
                  className="text-gray-700 hover:text-blue-600 px-3 py-2 text-sm font-medium"
                >
                  Auctions
                </Link>
                <Link
                  to="/create"
                  className="text-gray-700 hover:text-blue-600 px-3 py-2 text-sm font-medium"
                >
                  Create Auction
                </Link>
              </div>
            </div>
          </div>
        </div>
      </nav>

      {/* Page Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
}

/**
 * Main App component with router configuration
 */
function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<AuctionListPage />} />
          <Route path="/create" element={<CreateAuctionPage />} />
          <Route path="/auctions/:id" element={<AuctionDetailPage />} />
          <Route path="/auctions/:id/subscribe" element={<WebSocketPage />} />
        </Routes>
      </Layout>
      <Toaster position="top-right" />
    </BrowserRouter>
  );
}

export default App

