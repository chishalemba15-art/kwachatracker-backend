'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'

export default function NotificationsPage() {
  const [sending, setSending] = useState(false)
  const [form, setForm] = useState({
    title: '',
    body: '',
    target: 'all' as 'all' | 'active' | 'specific'
  })

  async function sendBroadcast() {
    if (!form.title || !form.body) {
      alert('Please fill in title and body')
      return
    }

    if (!confirm(`Send notification to ${form.target} users?`)) return

    setSending(true)
    try {
      await api.broadcast({
        title: form.title,
        body: form.body,
        target: form.target
      })
      alert('✅ Notification broadcast sent!')
      setForm({ title: '', body: '', target: 'all' })
    } catch (error) {
      alert('❌ Failed to send notification')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Push Notifications</h1>
        <p className="text-muted-foreground">Broadcast messages to users</p>
      </div>

      {/* Broadcast Form */}
      <Card>
        <CardHeader>
          <CardTitle>Send Broadcast</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Title</label>
            <input
              type="text"
              className="w-full p-2 border rounded-md"
              placeholder="e.g., Weekly Summary Available"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Message</label>
            <textarea
              className="w-full p-2 border rounded-md"
              rows={3}
              placeholder="e.g., Check your spending insights for this week!"
              value={form.body}
              onChange={(e) => setForm({ ...form, body: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Target</label>
            <select
              className="w-full p-2 border rounded-md"
              value={form.target}
              onChange={(e) => setForm({ ...form, target: e.target.value as any })}
            >
              <option value="all">All Users</option>
              <option value="active">Active Users (7 days)</option>
            </select>
          </div>

          <Button 
            onClick={sendBroadcast} 
            disabled={sending}
            className="w-full bg-blue-600 hover:bg-blue-700"
          >
            {sending ? 'Sending...' : 'Send Broadcast 📢'}
          </Button>
        </CardContent>
      </Card>

      {/* Preview */}
      <Card>
        <CardHeader>
          <CardTitle>Preview</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg p-4 bg-gray-50">
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0">
                <div className="h-10 w-10 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold">
                  K
                </div>
              </div>
              <div className="flex-1">
                <p className="font-semibold">{form.title || 'Notification Title'}</p>
                <p className="text-sm text-gray-600">{form.body || 'Notification message will appear here...'}</p>
                <p className="text-xs text-gray-400 mt-1">Just now</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
