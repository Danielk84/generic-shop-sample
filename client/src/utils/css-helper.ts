export const getCssVar = (cssVar: string) => (cssVar === "none" ? "none" : `var(${cssVar})`)

export const getCssUrl = (cssUrl?: string) => (cssUrl ? `url(${cssUrl})` : undefined)