import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from '@/components/providers/theme-provider';
import { Toaster } from "@/components/ui/toaster";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "FlowForge - Workflow Automation Platform",
  description: "Modern workflow automation platform with visual node-based editor",
  keywords: ["workflow", "automation", "n8n", "node-based", "visual programming"],
  authors: [{ name: "FlowForge Team" }],

  openGraph: {
    title: "FlowForge - Workflow Automation Platform",
    description: "Modern workflow automation platform with visual node-based editor",
    siteName: "FlowForge",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "FlowForge - Workflow Automation Platform",
    description: "Modern workflow automation platform with visual node-based editor",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${inter.variable} antialiased`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          {children}
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}