import { theme } from 'ant-design-vue'

const fontStack =
  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif"

/* ------------------------------------------------------------------ */
/*  Light theme + emerald accent                                       */
/*  bg-base: #F8FAFC  |  bg-surface: #FFFFFF  |  accent: #00D9A5       */
/* ------------------------------------------------------------------ */

export const customTheme = {
  token: {
    colorPrimary: '#00D9A5',
    colorInfo: '#60A5FA',
    colorSuccess: '#00D9A5',
    colorWarning: '#F59E0B',
    colorError: '#EF4444',
    colorLink: '#00D9A5',
    borderRadius: 12,
    borderRadiusLG: 14,
    borderRadiusSM: 8,
    colorBgLayout: '#F8FAFC',
    colorBgContainer: '#FFFFFF',
    colorBgElevated: '#F1F5F9',
    fontFamily: fontStack,
  },
  algorithm: theme.defaultAlgorithm,
}
