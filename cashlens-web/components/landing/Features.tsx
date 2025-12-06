"use client"

import { motion } from "framer-motion"
import { BarChart3, FileText, Lock, Zap } from "lucide-react"

const features = [
  {
    title: "Real-time Analytics",
    description: "Monitor your cash flow as it happens. Get instant insights into your financial health.",
    icon: BarChart3,
    className: "md:col-span-2",
  },
  {
    title: "Smart Categorization",
    description: "AI-powered transaction tagging saves you hours of manual data entry every month.",
    icon: Zap,
    className: "md:col-span-1",
  },
  {
    title: "Bank-Grade Security",
    description: "Your data is encrypted with AES-256. We never store your bank credentials.",
    icon: Lock,
    className: "md:col-span-1",
  },
  {
    title: "Automated Reports",
    description: "Generate professional financial reports for investors and stakeholders in one click.",
    icon: FileText,
    className: "md:col-span-2",
  },
]

export function Features() {
  return (
    <section id="features" className="py-24 bg-slate-50">
      <div className="container mx-auto px-4 md:px-6">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl mb-4">
            Everything you need to manage your finances
          </h2>
          <p className="text-lg text-muted-foreground">
            Stop wrestling with spreadsheets. Cashlens brings all your financial data into one beautiful dashboard.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {features.map((feature, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              className={`group relative overflow-hidden rounded-2xl border bg-white p-8 shadow-sm hover:shadow-md transition-shadow ${feature.className}`}
            >
              <div className="absolute top-0 right-0 -mt-4 -mr-4 h-24 w-24 rounded-full bg-primary/5 group-hover:bg-primary/10 transition-colors" />
              
              <div className="relative">
                <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <feature.icon className="h-6 w-6" />
                </div>
                <h3 className="mb-2 text-xl font-semibold">{feature.title}</h3>
                <p className="text-muted-foreground">{feature.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
