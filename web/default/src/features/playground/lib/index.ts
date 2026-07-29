/*
Copyright (C) 2023-2026 OpenFastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@example.com
*/
export * from './message-utils'
export * from './payload-builder'
export * from './storage'
export * from './message-styles'

// Re-export additional helpers consumed via '@/features/playground/lib'
export {
  FALLBACK_ERROR_CONTENT,
  MODEL_PRICING_SETTINGS_PATH,
  isAdminRole,
  isErrorMessage,
  getMessageErrorState,
} from './message/message-error-utils'
export {
  getChatMessageRenderState,
  getEditingMessageContent,
  getPreviousUserMessage,
  appendUserMessagePair,
  applyMessageEdit,
  createRegeneratedMessages,
  removeMessageByKey,
} from './message/conversation-message-utils'
export { getMessageContentState } from './message/message-content-utils'
export {
  getMessageActionState,
  getMessageActionsVisibilityClass,
} from './message/message-action-utils'
export { getMessageEditorState } from './message/message-editor-utils'
export {
  type MessageAlignment,
  getMessageAlignment,
  getMessageAlignmentClass,
} from './message/message-layout-utils'
export {
  ATTACHMENT_ACTIONS,
  getAttachmentActionNotice,
  getSearchActionNotice,
} from './input/input-tool-utils'
export {
  getInputControlState,
  getSubmittableInputText,
} from './input/input-control-utils'
export {
  getGroupFallback,
  getModelFallback,
  getOptionLoadErrorMessage,
  shouldClearModelForGroup,
} from './options/playground-option-utils'
