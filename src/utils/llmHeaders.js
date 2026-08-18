// Builds the request headers shared by all three LLM features (Chatbot, Who to Sign,
// Scout Report): the existing BYOK API key header, plus the optional configurable
// endpoint/model headers (map decision, see .scratch/llm-refinements/issues/
// 01-configurable-llm-endpoint-and-model.md). Omits the base-url/model headers
// entirely when unset, rather than sending them empty, so the backend's own
// per-field-default fallback behavior is exercised the same way either way.
export function llmRequestHeaders(uiStore) {
  const headers = {
    'Content-Type': 'application/json',
    'X-OpenAI-Api-Key': uiStore.openaiApiKey,
  }
  if (uiStore.openaiBaseUrl) headers['X-OpenAI-Base-URL'] = uiStore.openaiBaseUrl
  if (uiStore.openaiModel) headers['X-OpenAI-Model'] = uiStore.openaiModel
  return headers
}
