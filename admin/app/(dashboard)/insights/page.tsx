'use client'

import { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'
import { formatRelativeTime, truncate } from '@/lib/utils'
import { Brain, Zap, TrendingUp, DollarSign } from 'lucide-react'

interface Insight {
  id: string
  user_id: string
  type: string
  content: string
  generated_at: string
  delivered: boolean
}

export default function InsightsPage() {
  const [insights, setInsights] = useState<Insight[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [triggering, setTriggering] = useState(false)

  useEffect(() => {
    loadInsights()
  }, [page])

  async function loadInsights() {
    setLoading(true)
    try {
      const data = await api.getInsights({ page, limit: 50 })
      setInsights(data.insights || [])
      setTotal(data.total || 0)
    } catch (error) {
      console.error('Failed to load insights:', error)
    } finally {
      setLoading(false)
    }
  }

  async function triggerAnalysis() {
    if (!confirm('Trigger AI analysis for all users? This will use Gemini API credits.')) return
    
    setTriggering(true)
    try {
      await api.triggerInsights()
      alert('✅ AI analysis triggered! Check back in a few minutes.')
      setTimeout(loadInsights, 2000)
    } catch (error) {
      alert('❌ Failed to trigger analysis')
    } finally {
      setTriggering(false)
    }
  }

  const insightsToday = insights.filter(i => {
    const today = new Date().toDateString()
    return new Date(i.generated_at).toDateString() === today
  }).length

  return (
    <div className="p-8 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">AI Insights</h1>
          <p className="text-muted-foreground">Gemini-powered spending analysis</p>
        </div>
        <Button 
          onClick={triggerAnalysis} 
          disabled={triggering}
          className="bg-purple-600 hover:bg-purple-700"
        >
          <Zap className="mr-2 h-4 w-4" />
          {triggering ? 'Triggering...' : 'Trigger Analysis'}
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Insights</CardTitle>
            <Brain className="h-4 w-4 text-purple-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{total}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Generated Today</CardTitle>
            <TrendingUp className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{insightsToday}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Delivered</CardTitle>
            <Zap className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {insights.filter(i => i.delivered).length}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Est. API Cost</CardTitle>
            <DollarSign className="h-4 w-4 text-orange-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(insightsToday * 0.003).toFixed(3)}
            </div>
            <p className="text-xs text-muted-foreground">Today</p>
          </CardContent>
        </Card>
      </div>

      {/* Insights Table */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Insights</CardTitle>
          <CardDescription>
            AI-generated spending patterns and recommendations
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">Loading...</div>
          ) : insights.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No insights yet. Click "Trigger Analysis" to generate insights.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3 font-medium">Generated</th>
                    <th className="text-left p-3 font-medium">User</th>
                    <th className="text-left p-3 font-medium">Type</th>
                    <th className="text-left p-3 font-medium">Content</th>
                    <th className="text-left p-3 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {insights.map((insight) => (
                    <tr key={insight.id} className="border-b hover:bg-gray-50">
                      <td className="p-3 text-sm">
                        {formatRelativeTime(insight.generated_at)}
                      </td>
                      <td className="p-3">
                        <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                          {insight.user_id.slice(0, 8)}...
                        </code>
                      </td>
                      <td className="p-3">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                          {insight.type.replace('_', ' ')}
                        </span>
                      </td>
                      <td className="p-3 max-w-md">
                        <p className="text-sm truncate">{insight.content}</p>
                      </td>
                      <td className="p-3">
                        {insight.delivered ? (
                          <span className="text-green-600">✓ Delivered</span>
                        ) : (
                          <span className="text-yellow-600">⏳ Pending</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {/* Pagination */}
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
