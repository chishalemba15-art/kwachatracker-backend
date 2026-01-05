import Link from 'next/link'
import { LayoutDashboard, Brain, Users, Bell, CreditCard, BarChart3 } from 'lucide-react'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-6">
          <h2 className="text-2xl font-bold">Kwacha Tracker</h2>
          <p className="text-sm text-gray-400">Admin Dashboard</p>
        </div>

        <nav className="flex-1 px-4 space-y-1">
          <Link
            href="/"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition"
          >
            <LayoutDashboard className="h-5 w-5" />
            <span>Dashboard</span>
          </Link>

          <Link
            href="/insights"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition bg-purple-600"
          >
            <Brain className="h-5 w-5" />
            <span>AI Insights</span>
          </Link>

          <Link
            href="/users"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition"
          >
            <Users className="h-5 w-5" />
            <span>Users</span>
          </Link>

          <Link
            href="/notifications"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition"
          >
            <Bell className="h-5 w-5" />
            <span>Notifications</span>
          </Link>

          <Link
            href="/transactions"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition"
          >
            <CreditCard className="h-5 w-5" />
            <span>Transactions</span>
          </Link>

          <Link
            href="/analytics"
            className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition"
          >
            <BarChart3 className="h-5 w-5" />
            <span>Analytics</span>
          </Link>
        </nav>

        <div className="p-4 border-t border-gray-800">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-blue-600 flex items-center justify-center font-bold">
              A
            </div>
            <div>
              <p className="text-sm font-medium">Admin</p>
              <p className="text-xs text-gray-400">admin@kwachatracker.com</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        {/* Header */}
        <header className="bg-white border-b px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm text-gray-500">Welcome back</h3>
              <h2 className="text-xl font-semibold">Admin Dashboard</h2>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-500">
                API: {' '}
                <span className="text-green-600 font-medium">Online</span>
              </span>
            </div>
          </div>
        </header>

        {/* Page Content */}
        {children}
      </main>
    </div>
  )
}
