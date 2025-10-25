import { render, screen } from '@testing-library/react'
import { designSystem, themes } from '@/styles/design-system'

// 测试设计系统配置
describe('Design System', () => {
    it('should have correct color system', () => {
        expect(designSystem.colors.primary[500]).toBe('#1890ff')
        expect(designSystem.colors.success[500]).toBe('#52c41a')
        expect(designSystem.colors.warning[500]).toBe('#faad14')
        expect(designSystem.colors.error[500]).toBe('#f5222d')
    })

    it('should have correct typography system', () => {
        expect(designSystem.typography.fontSize.base).toBe('16px')
        expect(designSystem.typography.fontSize.lg).toBe('18px')
        expect(designSystem.typography.fontSize.xl).toBe('20px')
        expect(designSystem.typography.fontWeight.bold).toBe(700)
    })

    it('should have correct spacing system', () => {
        expect(designSystem.spacing[1]).toBe('4px')
        expect(designSystem.spacing[2]).toBe('8px')
        expect(designSystem.spacing[4]).toBe('16px')
        expect(designSystem.spacing[8]).toBe('32px')
    })

    it('should have correct border radius system', () => {
        expect(designSystem.borderRadius.sm).toBe('2px')
        expect(designSystem.borderRadius.base).toBe('4px')
        expect(designSystem.borderRadius.md).toBe('6px')
        expect(designSystem.borderRadius.lg).toBe('8px')
    })

    it('should have correct shadow system', () => {
        expect(designSystem.boxShadow.sm).toBe('0 1px 2px 0 rgba(0, 0, 0, 0.05)')
        expect(designSystem.boxShadow.base).toBe('0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)')
        expect(designSystem.boxShadow.lg).toBe('0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)')
    })

    it('should have correct animation system', () => {
        expect(designSystem.animation.duration.fast).toBe('0.1s')
        expect(designSystem.animation.duration.normal).toBe('0.2s')
        expect(designSystem.animation.duration.slow).toBe('0.3s')
    })

    it('should have correct breakpoints system', () => {
        expect(designSystem.breakpoints.xs).toBe('480px')
        expect(designSystem.breakpoints.sm).toBe('576px')
        expect(designSystem.breakpoints.md).toBe('768px')
        expect(designSystem.breakpoints.lg).toBe('992px')
    })

    it('should have correct z-index system', () => {
        expect(designSystem.zIndex.base).toBe(0)
        expect(designSystem.zIndex.dropdown).toBe(1000)
        expect(designSystem.zIndex.modal).toBe(1400)
        expect(designSystem.zIndex.tooltip).toBe(1800)
    })

    it('should have correct component sizes', () => {
        expect(designSystem.sizes.button.sm.height).toBe('24px')
        expect(designSystem.sizes.button.md.height).toBe('32px')
        expect(designSystem.sizes.button.lg.height).toBe('40px')

        expect(designSystem.sizes.input.sm.height).toBe('24px')
        expect(designSystem.sizes.input.md.height).toBe('32px')
        expect(designSystem.sizes.input.lg.height).toBe('40px')
    })

    it('should have light and dark themes', () => {
        expect(themes.light.mode).toBe('light')
        expect(themes.dark.mode).toBe('dark')

        expect(themes.light.colors.text.primary).toBe('#262626')
        expect(themes.dark.colors.text.primary).toBe('#ffffff')

        expect(themes.light.colors.background.primary).toBe('#ffffff')
        expect(themes.dark.colors.background.primary).toBe('#141414')
    })
})
