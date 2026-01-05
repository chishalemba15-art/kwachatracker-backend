'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function AnalyticsPage() {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Analytics</h1>
        <p className="text-muted-foreground">Detailed reports and insights</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Top Categories</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span>Airtime</span>
                <span className="font-semibold">45%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Data</span>
                <span className="font-semibold">30%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Transfers</span>
                <span className="font-semibold">15%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Payments</span>
                <span className="font-semibold">10%</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Operator Market Share</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span>Airtel</span>
                <span className="font-semibold">40%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>MTN</span>
                <span className="font-semibold">35%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Zamtel</span>
                <span className="font-semibold">20%</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Zedmobile</span>
                <span className="font-semibold">5%</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue Insights</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            Advanced analytics charts will be displayed here using Recharts.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
