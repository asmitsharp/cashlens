"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"
import { motion } from "framer-motion"
import { ArrowRight, BarChart3, PieChart, TrendingUp } from "lucide-react"

export function Hero() {
  return (
    <section className="relative overflow-hidden pt-32 pb-16 md:pt-48 md:pb-32">
      {/* Background Gradients */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[500px] bg-blue-500/20 rounded-full blur-[120px] -z-10" />
      <div className="absolute bottom-0 right-0 w-[800px] h-[600px] bg-purple-500/10 rounded-full blur-[100px] -z-10" />

      <div className="container mx-auto px-4 md:px-6">
        <div className="flex flex-col items-center text-center gap-8 max-w-4xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="inline-flex items-center rounded-full border border-primary/10 bg-primary/5 px-3 py-1 text-sm font-medium text-primary backdrop-blur-sm"
          >
            <span className="flex h-2 w-2 rounded-full bg-green-500 mr-2" />
            Now available for early access
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="text-4xl md:text-6xl lg:text-7xl font-bold tracking-tight"
          >
            Master Your Cash Flow with{" "}
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-purple-600">
              Intelligent Analytics
            </span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="text-xl text-muted-foreground max-w-2xl"
          >
            Cashlens gives Indian SMBs crystal-clear visibility into their finances. 
            Track expenses, forecast cash flow, and make data-driven decisions in seconds.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="flex flex-col sm:flex-row gap-4 w-full sm:w-auto"
          >
            <Link href="/sign-up">
              <Button size="lg" className="w-full sm:w-auto h-12 px-8 text-base gap-2">
                Start for Free <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Link href="#features">
              <Button variant="outline" size="lg" className="w-full sm:w-auto h-12 px-8 text-base">
                View Demo
              </Button>
            </Link>
          </motion.div>
        </div>

        {/* Dashboard Preview */}
        <motion.div
          initial={{ opacity: 0, y: 100, rotateX: 20 }}
          animate={{ opacity: 1, y: 0, rotateX: 0 }}
          transition={{ duration: 0.8, delay: 0.4, type: "spring" }}
          className="mt-16 relative mx-auto max-w-5xl perspective-1000"
        >
          <div className="relative rounded-xl border bg-white/50 backdrop-blur-xl shadow-2xl overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-tr from-white/50 via-transparent to-transparent z-10 pointer-events-none" />
            
            {/* Mock UI Header */}
            <div className="h-12 border-b bg-white/80 flex items-center px-4 gap-2">
              <div className="flex gap-1.5">
                <div className="w-3 h-3 rounded-full bg-red-400" />
                <div className="w-3 h-3 rounded-full bg-yellow-400" />
                <div className="w-3 h-3 rounded-full bg-green-400" />
              </div>
              <div className="ml-4 h-6 w-64 rounded-md bg-gray-100" />
            </div>

            {/* Mock UI Content */}
            <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6 bg-gray-50/50">
              {/* Card 1 */}
              <div className="rounded-xl bg-white p-6 shadow-sm border">
                <div className="flex items-center justify-between mb-4">
                  <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center text-blue-600">
                    <TrendingUp className="h-5 w-5" />
                  </div>
                  <span className="text-sm text-green-600 font-medium">+12.5%</span>
                </div>
                <div className="h-4 w-24 bg-gray-100 rounded mb-2" />
                <div className="h-8 w-32 bg-gray-200 rounded" />
              </div>

              {/* Card 2 */}
              <div className="rounded-xl bg-white p-6 shadow-sm border">
                <div className="flex items-center justify-between mb-4">
                  <div className="h-10 w-10 rounded-lg bg-purple-100 flex items-center justify-center text-purple-600">
                    <BarChart3 className="h-5 w-5" />
                  </div>
                  <span className="text-sm text-red-600 font-medium">-2.4%</span>
                </div>
                <div className="h-4 w-24 bg-gray-100 rounded mb-2" />
                <div className="h-8 w-32 bg-gray-200 rounded" />
              </div>

              {/* Card 3 */}
              <div className="rounded-xl bg-white p-6 shadow-sm border">
                <div className="flex items-center justify-between mb-4">
                  <div className="h-10 w-10 rounded-lg bg-orange-100 flex items-center justify-center text-orange-600">
                    <PieChart className="h-5 w-5" />
                  </div>
                  <span className="text-sm text-gray-500 font-medium">Now</span>
                </div>
                <div className="h-4 w-24 bg-gray-100 rounded mb-2" />
                <div className="h-8 w-32 bg-gray-200 rounded" />
              </div>

              {/* Chart Area */}
              <div className="md:col-span-2 rounded-xl bg-white p-6 shadow-sm border h-64 flex items-end gap-4 pb-4 px-4">
                 {[40, 65, 45, 80, 55, 70, 90].map((h, i) => (
                   <div key={i} className="flex-1 bg-blue-500/10 rounded-t-md relative group overflow-hidden">
                     <motion.div 
                       initial={{ height: 0 }}
                       animate={{ height: `${h}%` }}
                       transition={{ duration: 1, delay: 0.8 + (i * 0.1) }}
                       className="absolute bottom-0 w-full bg-blue-600 rounded-t-md"
                     />
                   </div>
                 ))}
              </div>

              {/* Recent Activity */}
              <div className="rounded-xl bg-white p-6 shadow-sm border h-64">
                <div className="space-y-4">
                  {[1, 2, 3, 4].map((i) => (
                    <div key={i} className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-full bg-gray-100" />
                      <div className="flex-1">
                        <div className="h-3 w-20 bg-gray-100 rounded mb-1" />
                        <div className="h-2 w-12 bg-gray-50 rounded" />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
