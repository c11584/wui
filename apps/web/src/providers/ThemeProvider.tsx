import { ConfigProvider, App as AntApp, theme as antTheme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import { useLayoutEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useThemeStore } from '../stores/theme'

interface ThemeProviderProps {
  children: React.ReactNode
}

export default function ThemeProvider({ children }: ThemeProviderProps) {
  const { i18n } = useTranslation()
  const { theme } = useThemeStore()

  const locale = i18n.language === 'zh' ? zhCN : enUS

  useLayoutEffect(() => {
    if (theme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark')
    } else {
      document.documentElement.removeAttribute('data-theme')
    }
  }, [theme])

  return (
    <ConfigProvider
      locale={locale}
      theme={{
        algorithm: theme === 'dark' ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
        token: {
          colorPrimary: '#6366f1',
          borderRadius: 6,
        },
      }}
    >
      <AntApp>{children}</AntApp>
    </ConfigProvider>
  )
}
