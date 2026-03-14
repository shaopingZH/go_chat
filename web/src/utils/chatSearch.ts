export interface SearchResultLike {
  id: number
}

export interface SearchResponseGuardInput {
  requestId: number
  activeRequestId: number
  requestKeyword: string
  currentKeyword: string
  requestKey: string
  currentKey: string
}

export function buildSearchTargetKey(results: SearchResultLike[], index: number): string {
  const target = index >= 0 && index < results.length ? results[index] : undefined
  return `${index}:${target?.id ?? ''}`
}

export function shouldApplySearchResponse(input: SearchResponseGuardInput): boolean {
  return input.requestId === input.activeRequestId &&
    input.requestKeyword === input.currentKeyword.trim() &&
    input.requestKey === input.currentKey
}
