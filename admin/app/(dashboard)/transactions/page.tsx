'use client'

import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'
import { formatDate, formatCurrency } from '@/lib/utils'

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  useEffect(() => {
    loadTransactions()
  }, [page])

  async function loadTransactions() {
    setLoading(true)
    try {
      const data = await api.getTransactions({ page, limit: 50 })
      setTransactions(data.transactions || [])
      setTotal(data.total || 0)
    } catch (error) {
      console.error('Failed to load transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Transactions</h1>
        <p className="text-muted-foreground">Browse all user transactions</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{total.toLocaleString()} Total Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">Loading...</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3">Date</th>
                    <th className="text-left p-3">User</th>
                    <th className="text-left p-3">Type</th>
                    <th className="text-left p-3">Category</th>
                    <th className="text-left p-3">Operator</th>
                    <th className="text-right p-3">Amount</th>
                    <th className="text-left p-3">Description</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions.map((txn) => (
                    <tr key={txn.id} className="border-b hover:bg-gray-50">
                      <td className="p-3 text-sm">{formatDate(txn.date)}</td>
                      <td className="p-3">
                        <code className="text-xs">{txn.user_id.slice(0, 8)}...</code>
                      </td>
                      <td className="p-3">
                        {txn.type === 'INCOME' ? (
                          <span className="text-green-600">↓ Income</span>
                        ) : (
                          <span className="text-red-600">↑ Expense</span>
                        )}
                      </td>
                      <td className="p-3">
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-gray-100">
                          {txn.category}
                        </span>
                      </td>
                      <td className="p-3 text-sm">{txn.operator}</td>
                      <td className="p-3 text-right font-medium">
                        {formatCurrency(txn.amount)}
                      </td>
                      <td className="p-3 text-sm text-gray-600 max-w-xs truncate">
                        {txn.description || '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              <div className="mt-6 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Showing {Math.min((page - 1) * 50 + 1, total)} to {Math.min(page * 50, total)} of {total}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setPage(p => p + 1)}
                    disabled={page * 50 >= total}
                  >
                    Next
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
