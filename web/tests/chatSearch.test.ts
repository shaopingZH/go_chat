import test from 'node:test'
import assert from 'node:assert/strict'

import { buildSearchTargetKey, shouldApplySearchResponse } from '../src/utils/chatSearch.ts'

test('buildSearchTargetKey changes when the selected result id changes at same index', () => {
  const first = buildSearchTargetKey([{ id: 101 }, { id: 102 }], 0)
  const second = buildSearchTargetKey([{ id: 201 }, { id: 202 }], 0)

  assert.equal(first, '0:101')
  assert.equal(second, '0:201')
  assert.notEqual(first, second)
})

test('buildSearchTargetKey is empty when index is invalid', () => {
  assert.equal(buildSearchTargetKey([], -1), '-1:')
  assert.equal(buildSearchTargetKey([{ id: 1 }], 2), '2:')
})

test('shouldApplySearchResponse rejects stale request ids', () => {
  const shouldApply = shouldApplySearchResponse({
    requestId: 1,
    activeRequestId: 2,
    requestKeyword: 'father',
    currentKeyword: 'father',
    requestKey: 'private:2',
    currentKey: 'private:2',
  })

  assert.equal(shouldApply, false)
})

test('shouldApplySearchResponse rejects responses for outdated keyword or conversation', () => {
  assert.equal(
    shouldApplySearchResponse({
      requestId: 3,
      activeRequestId: 3,
      requestKeyword: "I'am your father",
      currentKeyword: '5',
      requestKey: 'private:2',
      currentKey: 'private:2',
    }),
    false,
  )

  assert.equal(
    shouldApplySearchResponse({
      requestId: 3,
      activeRequestId: 3,
      requestKeyword: 'father',
      currentKeyword: 'father',
      requestKey: 'private:2',
      currentKey: 'group:9',
    }),
    false,
  )
})

test('shouldApplySearchResponse accepts current request for same keyword and conversation', () => {
  const shouldApply = shouldApplySearchResponse({
    requestId: 4,
    activeRequestId: 4,
    requestKeyword: 'father',
    currentKeyword: ' father ',
    requestKey: 'private:2',
    currentKey: 'private:2',
  })

  assert.equal(shouldApply, true)
})
