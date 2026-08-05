/*
Copyright (C) 2023-2026 FastToken
*/
import { useCallback, useEffect, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { getUserModels, getUserGroups } from "./api"
import { getApiKeys } from "@/features/keys/api"
import { useAuthStore } from "@/stores/auth-store"
import { ROLE } from "@/lib/roles"
import { PlaygroundChat } from "./components/playground-chat"
import { PlaygroundInput } from "./components/playground-input"
import { usePlaygroundState, useChatHandler } from "./hooks"
import { createUserMessage, createLoadingAssistantMessage } from "./lib"
import type { Message as MessageType } from "./types"

const MIN_BALANCE_QUOTA = 1000

export function Playground() {
  const { t } = useTranslation()
  const {
    config, parameterEnabled, messages, models, groups,
    updateMessages, setModels, setGroups, updateConfig,
  } = usePlaygroundState()

  const { auth } = useAuthStore()
  const isAdmin = (auth.user?.role ?? 0) >= ROLE.ADMIN
  const userQuota = auth.user?.quota ?? 0

  const { sendChat, stopGeneration, isGenerating } = useChatHandler({
    config, parameterEnabled, onMessageUpdate: updateMessages,
  })

  const [editingMessageKey, setEditingMessageKey] = useState<string | null>(null)
  const [hasActiveTokens, setActiveTokens] = useState<boolean | null>(null)

  const { data: tokensData } = useQuery({
    queryKey: ["playground-user-tokens"],
    queryFn: async () => {
      try {
        const res = await getApiKeys({ p: 1, size: 5 })
        // getApiKeys 返回响应体本身：{ success, message, data?: { items, total, ... } }
        // 与 Keys 页面一致消费方式：res.data?.items
        const tokens = res?.data?.items ?? []
        // 只有“成功响应且列表为空”才表示用户确实没有令牌。
        // 其它任何状态（传输错误、success:false、结构异常）都视为“不确定”，
        // 绝不能因此锁定一个可能真实拥有令牌的用户。
        if (!res?.success || !Array.isArray(tokens)) return null
        return tokens.length > 0
      } catch {
        return null
      }
    },
    staleTime: 60_000,
  })

  useEffect(() => { if (tokensData !== undefined) setActiveTokens(tokensData) }, [tokensData])

  const { data: modelsData, isLoading: isLoadingModels } = useQuery({
    queryKey: ["playground-models"],
    queryFn: async () => { try { return await getUserModels() } catch (e) { toast.error(e instanceof Error ? e.message : t("Failed to load models")); return [] } },
  })

  const { data: groupsData } = useQuery({
    queryKey: ["playground-groups"],
    queryFn: async () => { try { return await getUserGroups() } catch (e) { toast.error(e instanceof Error ? e.message : t("Failed to load groups")); return [] } },
  })

  useEffect(() => {
    if (!modelsData) return
    setModels(modelsData)
    if (modelsData.length > 0 && !modelsData.some((m) => m.value === config.model)) updateConfig("model", modelsData[0].value)
  }, [modelsData, config.model, setModels, updateConfig])

  useEffect(() => {
    if (!groupsData || groupsData.length === 0) return
    setGroups(groupsData)
    if (!groupsData.some((g) => g.value === config.group)) {
      const fb = groupsData.find((g) => g.value === "auto")?.value ?? groupsData.find((g) => g.value === "default")?.value ?? groupsData[0].value
      updateConfig("group", fb)
    }
  }, [groupsData, setGroups, config.group, updateConfig])

  const validatePrerequisites = useCallback((): boolean => {
    if (userQuota <= MIN_BALANCE_QUOTA) {
      toast.error("\u65e0\u6cd5\u53d1\u8d77\u5bf9\u8bdd\uff1a\u94b1\u5305\u4f59\u989d\u4e0d\u8db3\u3002\u8bf7\u5148\u524d\u5f80\u5145\u503c\u540e\u518d\u8bd5\u3002", { duration: 5000, action: { label: "\u524d\u5f80\u5145\u503c", onClick: () => { window.location.href = "/wallet" } } })
      return false
    }
    if (hasActiveTokens === null) return true
    if (!hasActiveTokens) {
      toast.error("\u65e0\u6cd5\u53d1\u8d77\u5bf9\u8bdd\uff1a\u60a8\u8fd8\u6ca1\u6709\u521b\u5efa API \u4ee4\u724c\u3002\u8bf7\u521b\u5efa\u4ee4\u724c\u540d\u518d\u8bd5\u3002", { duration: 5000, action: { label: "\u521b\u5efa\u4ee4\u724c", onClick: () => { window.location.href = "/keys" } } })
      return false
    }
    return true
  }, [userQuota, hasActiveTokens])

  const handleSendMessage = (text: string) => {
    if (!validatePrerequisites()) return
    const userMsg = createUserMessage(text)
    const asstMsg = createLoadingAssistantMessage()
    updateMessages([...messages, userMsg, asstMsg])
    sendChat([...messages, userMsg, asstMsg])
  }

  const handleCopyMessage = (_m: MessageType) => {}
  const handleRegenerateMessage = (message: MessageType) => {
    if (!validatePrerequisites()) return
    const idx = messages.findIndex((m) => m.key === message.key)
    if (idx === -1) return
    updateMessages([...messages.slice(0, idx), createLoadingAssistantMessage()])
    sendChat([...messages.slice(0, idx), createLoadingAssistantMessage()])
  }

  const handleEditMessage = useCallback((m: MessageType) => setEditingMessageKey(m.key), [])
  const handleEditOpenChange = useCallback((open: boolean) => { if (!open) setEditingMessageKey(null) }, [])

  const applyEdit = useCallback((newContent: string, submit: boolean) => {
    if (!editingMessageKey) return
    const idx = messages.findIndex((m) => m.key === editingMessageKey)
    if (idx === -1) return
    const updated = messages.map((m) => m.key === editingMessageKey ? { ...m, versions: [{ ...m.versions[0], content: newContent }] } : m)
    setEditingMessageKey(null)
    if (!submit || updated[idx].from !== "user") { updateMessages(updated); return }
    if (!validatePrerequisites()) return
    const toSub = [...updated.slice(0, idx + 1), createLoadingAssistantMessage()]
    updateMessages(toSub)
    sendChat(toSub)
  }, [editingMessageKey, messages, updateMessages, sendChat, validatePrerequisites])

  const handleDeleteMessage = (m: MessageType) => updateMessages(messages.filter((x) => x.key !== m.key))

  return (
    <div className="relative flex size-full flex-col overflow-hidden">
      <div className="flex flex-1 flex-col overflow-hidden">
        <PlaygroundChat messages={messages} onCopyMessage={handleCopyMessage}
          onRegenerateMessage={handleRegenerateMessage} onEditMessage={handleEditMessage}
          onDeleteMessage={handleDeleteMessage} isGenerating={isGenerating}
          editingKey={editingMessageKey} onCancelEdit={handleEditOpenChange}
          onSaveEdit={(c) => applyEdit(c, false)} onSaveEditAndSubmit={(c) => applyEdit(c, true)}
        />
      </div>
      <div className="mx-auto w-full max-w-4xl">
        <PlaygroundInput disabled={isGenerating} groups={groups} groupValue={config.group}
          isGenerating={isGenerating} isModelLoading={isLoadingModels}
          modelValue={config.model} models={models} onGroupChange={(v) => updateConfig("group", v)}
          onModelChange={(v) => updateConfig("model", v)} onStop={stopGeneration}
          onSubmit={handleSendMessage} isAdmin={isAdmin}
        />
      </div>
    </div>
  )
}
