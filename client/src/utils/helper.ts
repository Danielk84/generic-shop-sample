export const getCssVar = (cssVar: string) => (cssVar === "none" ? "none" : `var(${cssVar})`)

export const getCssUrl = (cssUrl?: string) => (cssUrl ? `url(${cssUrl})` : undefined)

export function* range(start: number, end: number, step: number = 1) {
    for (let i = start; i < end; i += step) {
        yield i;
    }
}
