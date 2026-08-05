import { useTranslation } from 'react-i18next'
import type { TFunction } from 'react-i18next'
/*
Copyright (C) 2023-2026 FastToken

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

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { useState, useEffect } from 'react'
import {
  BookOpen,
  Zap,
  Key,
  Network,
  Shield,
  BarChart3,
  HelpCircle,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  Info,
  Sparkles,
  Layers,
  Users,
  Globe,
  Server,
  Terminal,
  DollarSign,
  Cpu,
  Download,
  Package,
  TerminalSquare,
} from 'lucide-react'
import { PublicLayout } from '@/components/layout'
import { api } from '@/lib/api'
import { CopyButton } from '@/components/copy-button'

interface DocSection {
  id: string
  icon: React.ComponentType<{ className?: string }>
  title: string
  content: React.ReactNode
}

interface QAItem {
  q: string
  a: string
}

interface ModelInfo {
  id: string
  object?: string
  owned_by?: string
}

interface PricingInfo {
  model_name: string
  model_ratio?: number
  prompt_ratio?: number
  completion_ratio?: number
  group?: string
}

const BASE_URL = 'https://fasttoken.example.com'

function buildDocSections(
  t: TFunction,
  models: ModelInfo[],
  _pricing: PricingInfo[],
): DocSection[] {
  return [
    // ===== 1. 产品介绍 =====
    {
      id: 'product-intro',
    icon: Sparkles,
    title: '产品介绍',
    content: (
      <div className="space-y-6">
        <p className="text-stone-muted leading-relaxed">{t('FastToken 是企业级 AI 统一网关，基于')}<strong>{t('中国电信天翼云极速推理')}</strong>{t('， 兼容 OpenAI SDK。通过一个标准 API 即可访问 DeepSeek、Qwen 等主流大模型，无需适配多个供应商。')}</p>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[
          { title: '多模型聚合', desc: '一个接口访问 DeepSeek、Qwen、Claude 等数十个模型', icon: Layers },
          { title: '按量付费', desc: 'Token 级用量追踪，缓存命中折扣，成本透明可控', icon: BarChart3 },
          { title: '企业级稳定', desc: '多通道负载均衡、自动故障切换，99.9% SLA 保障', icon: Shield },
        ].map((item, idx) => (
          <div key={idx} className="bg-card rounded-xl p-5 shadow-sm border border-stone-card
                          hover:shadow-md hover:-translate-y-1 transition-all duration-300">
            <div className="w-10 h-10 bg-accent-brand rounded-lg flex items-center justify-center mb-3">
              <item.icon className="w-5 h-5 text-white" />
            </div>
            <h4 className="text-base font-serif font-bold text-stone-text mb-2">{item.title}</h4>
            <p className="text-sm text-stone-muted leading-relaxed">{item.desc}</p>
          </div>
        ))}
      </div>

      <div className="bg-accent-brand/5 border border-accent-brand/20 rounded-xl p-5">
        <p className="text-sm text-stone-text leading-relaxed">{t('所有模型均通过')}<strong>{t('中国电信天翼云')}</strong>{t('极速推理服务提供，享受运营商级网络低延迟优势。')}</p>
      </div>
    </div>
  ),
},

// ===== 2. 快速开始 =====
{
  id: 'quick-start',
  icon: Zap,
  title: '快速开始',
  content: (
    <div className="space-y-6">
      <p className="text-stone-muted leading-relaxed">{t('三步接入 FastToken，5 分钟内发出第一个 API 请求。')}</p>

    <div className="space-y-4">
      {[
        {
          step: '1',
          title: '注册账号',
          desc: '访问 FastToken 控制台，使用邮箱注册。注册即享免费额度，无需绑定支付方式。',
        },
        {
          step: '2',
          title: '创建 API Key',
          desc: '登录后进入「API Key 管理」，点击「创建 API Key」，复制生成的 Key（格式：sk-...）。可为 Key 设置模型限制和额度。',
        },
        {
          step: '3',
          title: '发起第一个请求',
          desc: '将以下代码中的 YOUR_API_KEY 替换为你的 Key，即可调用 AI 模型。支持 curl、Python、Node.js 等多种方式。',
        },
      ].map((item, idx) => (
        <div key={idx} className="bg-card rounded-xl p-6 shadow-sm border border-stone-card
                        hover:shadow-md hover:-translate-y-1 transition-all duration-300">
          <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center space-x-2">
            <span className="w-8 h-8 bg-primary text-primary-foreground rounded-lg flex items-center justify-center text-sm font-bold">{item.step}</span>
            <span>{item.title}</span>
          </h4>
          <p className="text-stone-muted leading-relaxed">{item.desc}</p>
        </div>
      ))}
    </div>

    <div className="relative">
      <pre className="bg-stone-bg rounded-lg p-4 text-xs font-mono text-stone-text overflow-x-auto leading-relaxed">
{`from openai import OpenAI

client = OpenAI(
    base_url="https://fasttoken.example.com/v1",
    api_key="YOUR_API_KEY"
)

response = client.chat.completions.create(
    model="deepseek-v4-pro",
    messages=[{"role": "user", "content": "你好，请介绍一下自己"}]
)
print(response.choices[0].message.content)`}
      </pre>
      <div className="absolute top-2 right-2">
        <CopyButton
          value={`from openai import OpenAI\n\nclient = OpenAI(\n    base_url="https://fasttoken.example.com/v1",\n    api_key="YOUR_API_KEY"\n)\n\nresponse = client.chat.completions.create(\n    model="deepseek-v4-pro",\n    messages=[{"role": "user", "content": "你好，请介绍一下自己"}]\n)\nprint(response.choices[0].message.content)`}
          variant="outline"
          className="border-stone-card bg-card size-7"
          iconClassName="size-3"
        />
      </div>
    </div>
  </div>
),
},

// ===== 3. API 参考 =====
{
  id: 'api-reference',
  icon: Terminal,
  title: 'API 参考',
  content: (
    <div className="space-y-6">
      {/* Endpoint */}
      <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
        <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
          <Server className="w-5 h-5 text-accent-brand" />{t('API 端点')}</h4>
        <p className="text-stone-muted text-sm mb-3">{t('Chat Completions（OpenAI 兼容）：')}</p>
        <div className="bg-stone-bg rounded-lg p-4 flex items-center justify-between gap-2">
          <code className="text-stone-text font-mono text-sm break-all">
            {BASE_URL}/v1/chat/completions
          </code>
          <CopyButton
            value={`${BASE_URL}/v1/chat/completions`}
            variant="outline"
            className="border-stone-card bg-card size-8 shrink-0"
            iconClassName="size-3.5"
          />
        </div>
    </div>

    {/* Authentication */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
        <Key className="w-5 h-5 text-accent-brand" />{t('认证方式')}</h4>
      <p className="text-stone-muted text-sm mb-3">{t('在')}<code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs">Authorization</code>{t('请求头中携带 API Key：')}</p>
      <div className="bg-stone-bg rounded-lg p-4">
        <code className="text-stone-text font-mono text-sm">
          Authorization: Bearer YOUR_API_KEY
        </code>
      </div>
    </div>

    {/* Python Example */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
        <Cpu className="w-5 h-5 text-accent-brand" />{t('Python SDK 示例')}</h4>
      <div className="relative">
        <pre className="bg-stone-bg rounded-lg p-4 text-xs font-mono text-stone-text overflow-x-auto leading-relaxed">
{`from openai import OpenAI

client = OpenAI(
    base_url="https://fasttoken.example.com/v1",
    api_key="YOUR_API_KEY"
)

# 基础对话
response = client.chat.completions.create(
    model="deepseek-v4-pro",
    messages=[
        {"role": "system", "content": "你是一个数学老师"},
        {"role": "user", "content": "什么是傅里叶变换？"}
    ],
    temperature=0.7,
    max_tokens=1000
)

# 流式输出
stream = client.chat.completions.create(
    model="deepseek-v4-pro",
    messages=[{"role": "user", "content": "写一首关于春天的诗"}],
    stream=True
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")`}
        </pre>
        <div className="absolute top-2 right-2">
          <CopyButton
            value={`from openai import OpenAI\n\nclient = OpenAI(\n    base_url="https://fasttoken.example.com/v1",\n    api_key="YOUR_API_KEY"\n)\n\nresponse = client.chat.completions.create(\n    model="deepseek-v4-pro",\n    messages=[\n        {"role": "system", "content": "你是一个数学老师"},\n        {"role": "user", "content": "什么是傅里叶变换？"}\n    ],\n    temperature=0.7,\n    max_tokens=1000\n)`}
            variant="outline"
            className="border-stone-card bg-card size-7"
            iconClassName="size-3"
          />
        </div>
      </div>
    </div>

    {/* cURL Example */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
        <Terminal className="w-5 h-5 text-accent-brand" />{t('cURL 示例')}</h4>
      <div className="relative">
        <pre className="bg-stone-bg rounded-lg p-4 text-xs font-mono text-stone-text overflow-x-auto leading-relaxed">
{`curl https://fasttoken.example.com/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [{"role": "user", "content": "你好！"}]
  }'`}
        </pre>
        <div className="absolute top-2 right-2">
          <CopyButton
            value={`curl https://fasttoken.example.com/v1/chat/completions \\\n  -H "Content-Type: application/json" \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -d '{\\n    "model": "deepseek-v4-pro",\\n    "messages": [{"role": "user", "content": "你好！"}]\\n  }'`}
            variant="outline"
            className="border-stone-card bg-card size-7"
            iconClassName="size-3"
          />
        </div>
      </div>
    </div>

    {/* Supported Endpoints */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3">{t('支持的接口')}</h4>
      <div className="space-y-2 text-sm">
        {[
          { name: 'Chat Completions', path: '/v1/chat/completions', desc: '聊天补全（支持 stream）' },
          { name: 'Models', path: '/v1/models', desc: '获取可用模型列表' },
          { name: 'Embeddings', path: '/v1/embeddings', desc: '文本向量嵌入' },
          { name: 'Images', path: '/v1/images/generations', desc: '图片生成' },
          { name: 'Audio', path: '/v1/audio/*', desc: '语音转文字 / 文字转语音' },
        ].map((ep, idx) => (
          <div key={idx} className="flex items-center justify-between py-1">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-accent-brand flex-shrink-0" />
              <span className="font-medium text-stone-text">{ep.name}</span>
              <span className="text-stone-muted">— {ep.desc}</span>
            </div>
            <code className="text-xs bg-stone-bg px-2 py-0.5 rounded font-mono text-stone-muted">{ep.path}</code>
          </div>
        ))}
      </div>
    </div>

    {/* 分组头 */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3">{t('自定义请求头')}</h4>
      <div className="space-y-2 text-sm text-stone-muted">
        <div className="flex items-center gap-3">
          <code className="bg-stone-bg px-2 py-0.5 rounded font-mono text-xs">X-FastToken-Group</code>
          <span>{t('分组名称，如 ')}<code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs">auto</code>、<code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs">vip</code></span>
        </div>
      </div>
    </div>

    {/* Error Codes */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3">{t('错误码')}</h4>
      <div className="space-y-2 text-sm">
        {[
          { code: '401', desc: 'API Key 无效或已过期' },
          { code: '402', desc: '余额不足，请充值' },
          { code: '403', desc: '无权访问该模型' },
          { code: '429', desc: '请求频率超限，请稍后重试' },
          { code: '500', desc: '服务器内部错误' },
          { code: '502', desc: '上游服务异常，自动重试中' },
        ].map((err, idx) => (
          <div key={idx} className="flex items-center gap-3">
            <code className="bg-accent-brand/10 text-accent-brand px-2 py-0.5 rounded font-mono text-xs w-10 text-center flex-shrink-0">
              {err.code}
            </code>
            <span className="text-stone-muted">{err.desc}</span>
          </div>
        ))}
      </div>
    </div>
  </div>
),
},

// ===== 可用模型 =====
{
  id: 'models',
  icon: Cpu,
  title: '可用模型',
  content: (
    <div className="space-y-6">
      <p className="text-stone-muted leading-relaxed">{t('以下模型通过 FastToken 网关提供，所有模型均基于')}<strong>{t('中国电信天翼云极速推理')}</strong>{t('。 在 API 请求的')}<code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs">model</code>{t('字段中使用模型 ID。')}</p>

    {/* 模型选择指南 */}
    <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
      <h4 className="text-lg font-serif font-bold text-stone-text mb-3">{t('模型选择指南')}</h4>
      <div className="space-y-3 text-sm">
        {[
          { scene: '代码生成 / 复杂推理', model: 'deepseek-v4-pro', desc: 'DeepSeek V4 Pro 旗舰推理能力' },
          { scene: '日常对话 / 翻译', model: 'deepseek-v3 / qwen3-max', desc: '平衡性能与成本' },
          { scene: '实时交互 / 极速响应', model: 'qwen-flash', desc: '最低延迟、最快响应' },
          { scene: '长文档 / 多语言', model: 'qwen3.5-397b-a17b', desc: '397B 参数超大规模' },
          { scene: '图片理解 / OCR', model: 'qwen3-vl-235b-a22b-instruct', desc: '多模态视觉理解' },
        ].map((item, idx) => (
          <div key={idx} className="flex items-start gap-3 py-2 border-b border-stone-card/60 last:border-b-0">
            <CheckCircle2 className="w-4 h-4 text-accent-brand mt-0.5 flex-shrink-0" />
            <div>
              <span className="font-medium text-stone-text">{item.scene}</span>
              <span className="text-stone-muted mx-1">→</span>
              <code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs font-mono text-accent-brand">{item.model}</code>
              <p className="text-stone-muted text-xs mt-0.5">{item.desc}</p>
            </div>
          </div>
        ))}
      </div>
    </div>

    {models.length > 0 ? (
      <div className="bg-card rounded-xl shadow-sm border border-stone-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-stone-card bg-stone-bg/50">
                <th className="px-4 py-3 text-left font-serif font-bold text-stone-text">{t('模型 ID')}</th>
                <th className="px-4 py-3 text-left font-serif font-bold text-stone-text">{t('供应商')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-stone-card/60">
              {models.map((model, idx) => (
                <tr key={idx} className="hover:bg-stone-bg/30 transition-colors">
                  <td className="px-4 py-2.5 font-mono text-xs text-stone-text">{model.id}</td>
                  <td className="px-4 py-2.5 text-stone-muted">{model.owned_by || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    ) : (
      <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card text-center text-stone-muted text-sm">
        <Info className="w-5 h-5 mx-auto mb-2 text-stone-muted/60" />{t('模型列表加载中...')}</div>
    )}

    <div className="bg-accent-brand/5 border border-accent-brand/20 rounded-xl p-5">
      <p className="text-sm text-stone-text">{t('调用')}<code className="bg-stone-bg px-1.5 py-0.5 rounded text-xs">GET /v1/models</code>{t('获取最新可用模型列表。 模型列表持续更新，新增或下线模型前会在控制台公告中通知。')}</p>
    </div>
  </div>
),
},

// ===== 5. 功能指南 =====
{
  id: 'feature-guide',
  icon: BookOpen,
  title: '功能指南',
  content: (
    <div className="space-y-4">
      {[
        {
          icon: Network,
          title: '模型管理',
          desc: '通过单一端点访问数十个模型，只需更改 model 参数即可切换。支持 OpenAI、Claude、Gemini、DeepSeek、Qwen 等主流模型。',
          models: ['DeepSeek V4 Pro / V3', 'Qwen 3.5 397B / Qwen 3 Max', 'Qwen Flash（极速）', 'Qwen VL（多模态）'],
        },
        {
          icon: Globe,
          title: '渠道管理',
          desc: '添加和管理上游 AI 供应商。智能路由、负载均衡、自动故障切换，保障服务高可用。',
          items: ['多供应商支持', '加权负载均衡', '失败自动重试', '优先级路由'],
        },
        {
          icon: Key,
          title: 'API Key 管理',
          desc: '创建和管理 API 密钥，精细控制权限和配额。支持环境变量存储，安全便捷。',
          items: ['一键生成 Key', '模型访问限制', '配额额度控制', '有效期设置'],
        },
        {
          icon: DollarSign,
          title: '会员与支付',
          desc: '支持支付宝和微信支付充值。Standard / Pro / Enterprise 三级会员，自动匹配折扣。',
          items: ['支付宝 / 微信支付', 'Standard / Pro / Enterprise', 'auto 分组自动折扣', '用量告警通知'],
        },
        {
          icon: Users,
          title: '企业团队管理',
          desc: '通过分组批量管理用户，差异化访问策略。子账号、统一账单、IP 白名单。',
          items: ['分组管理', '子账号额度分配', '模型访问控制', '角色权限管理'],
        },
        {
          icon: Shield,
          title: '安全设置',
          desc: '企业级安全防护：IP 白名单、访问日志审计、地域限制、HTTPS 加密传输。',
          items: ['IP 白名单', 'API 访问日志', '地域限制', '数据加密传输'],
        },
      ].map((feature, idx) => (
        <div key={idx} className="bg-card rounded-xl p-6 shadow-sm border border-stone-card
                        hover:shadow-md hover:-translate-y-1 transition-all duration-300">
          <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
            <feature.icon className="w-5 h-5 text-accent-brand" />
            <span>{feature.title}</span>
          </h4>
          <p className="text-stone-muted text-sm leading-relaxed mb-3">{feature.desc}</p>
          {'models' in feature && feature.models && (
            <div className="grid grid-cols-2 sm:grid-cols-2 gap-2">
              {feature.models.map((m, i) => (
                <div key={i} className="flex items-center gap-2 text-sm text-stone-muted">
                  <CheckCircle2 className="w-3.5 h-3.5 text-accent-brand flex-shrink-0" />
                  <span>{m}</span>
                </div>
              ))}
            </div>
          )}
          {'items' in feature && feature.items && (
            <ul className="space-y-1.5 text-sm text-stone-muted">
              {feature.items.map((item, i) => (
                <li key={i} className="flex items-center gap-2">
                  <ArrowRight className="w-3.5 h-3.5 text-accent-brand flex-shrink-0" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      ))}
    </div>
  ),
},

// ===== 6. 常见问题 =====
{
  id: 'faq',
  icon: HelpCircle,
  title: '常见问题',
  content: (
    <div className="space-y-4">
      {( [
        {
          q: '如何切换模型？',
          a: '只需更改 API 请求中的 model 参数。例如将 "model": "deepseek-v4-pro" 改为 "model": "qwen3-max"。调用 GET /v1/models 获取所有可用模型列表。',
        },
        {
          q: '支持哪些 AI 供应商？',
          a: 'OpenAI、DeepSeek、Qwen（通义千问）等数十个主流供应商，均通过中国电信天翼云极速推理提供。也支持自定义 HTTP 上游接入任何 OpenAI 兼容服务。',
        },
        {
          q: '如何计费？',
          a: '按量付费，精确到 Token 级别。输入和输出 Token 分开计费，支持缓存命中折扣（如 DeepSeek 缓存命中享折扣价格）。在控制台可实时查看用量和费用。',
        },
        {
          q: '速率限制是多少？',
          a: '默认每分钟 60 次请求（RPM）。企业用户可申请更高限额。超限时返回 429 状态码，建议客户端实现指数退避重试。',
        },
        {
          q: '请求失败怎么办？',
          a: 'FastToken 自动在其他可用渠道重试。当一个渠道失败时，会切换到其他渠道继续请求。可在控制台「日志」页面查看失败详情和重试记录。',
        },
        {
          q: '如何获取 API Key？',
          a: '登录控制台 →「API Key 管理」→「创建 API Key」→ 复制 Key（仅显示一次）。建议使用环境变量存储，不要硬编码在代码中。',
        },
        {
          q: '支持流式输出吗？',
          a: '支持。在请求中设置 stream: true 即可启用 SSE 流式输出，实时获取模型逐字生成的响应。',
        },
        {
          q: '如何联系技术支持？',
          a: '通过控制台提交工单，或发送邮件至 hello@fasttoken.example.com。工作日 9:00-18:00 提供技术支持。',
        },
      ] as QAItem[]).map((qa, idx) => (
        <div key={idx} className="bg-card rounded-xl p-6 shadow-sm border border-stone-card
                        hover:shadow-md hover:-translate-y-1 transition-all duration-300">
          <h4 className="text-base font-serif font-bold text-stone-text mb-3 flex items-start gap-2">
            <AlertCircle className="w-5 h-5 text-accent-brand mt-0.5 flex-shrink-0" />
            <span>{qa.q}</span>
          </h4>
          <p className="text-sm text-stone-muted leading-relaxed pl-7">{qa.a}</p>
        </div>
      ))}
    </div>
  ),
},

  // ===== 7. FastToken 技能 =====
  {
    id: 'skills',
    icon: Package,
    title: 'FastToken 技能',
    content: (
      <div className="space-y-6">
        <p className="text-stone-muted leading-relaxed">{t('FastToken 提供一套 AI 编程助手技能（Skill），安装后助手可获得网关管理、API 调用、用量与计费查询等能力，无需手动拼接请求。')}</p>

        {/* 下载 ZIP */}
        <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
          <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
            <Download className="w-5 h-5 text-accent-brand" />{t('下载技能包')}
          </h4>
          <p className="text-stone-muted text-sm mb-4">{t('包含 fasttoken 用户端技能，解压后即可装入助手技能目录。')}</p>
          <a
            href="/skills/fasttoken-skills.zip"
            className="inline-flex items-center gap-2 bg-primary text-primary-foreground px-5 py-3 rounded-lg font-medium hover:opacity-90 transition-all duration-200"
          >
            <Download className="w-4 h-4" />
            <span>{t('下载 fasttoken-skills.zip')}</span>
          </a>
        </div>

        {/* CLI 一键安装 */}
        <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
          <h4 className="text-lg font-serif font-bold text-stone-text mb-3 flex items-center gap-2">
            <TerminalSquare className="w-5 h-5 text-accent-brand" />{t('CLI 一键安装')}
          </h4>
          <p className="text-stone-muted text-sm mb-4">{t('下载安装脚本并运行，自动从官网拉取最新技能包并装入技能目录。')}</p>
          <div className="relative">
            <pre className="bg-stone-bg rounded-lg p-4 text-xs font-mono text-stone-text overflow-x-auto leading-relaxed">
{`# macOS / Linux
curl -fsSL https://fasttoken.example.com/skills/install-fasttoken-skills.py -o install.py
python3 install.py

# Windows (PowerShell)
Invoke-WebRequest https://fasttoken.example.com/skills/install-fasttoken-skills.py -OutFile install.py
python install.py`}
            </pre>
            <div className="absolute top-2 right-2">
              <CopyButton
                value={`curl -fsSL https://fasttoken.example.com/skills/install-fasttoken-skills.py -o install.py\npython3 install.py`}
                variant="outline"
                className="border-stone-card bg-card size-7"
                iconClassName="size-3"
              />
            </div>
          </div>
        </div>

        {/* 安装方式说明 */}
        <div className="bg-card rounded-xl p-6 shadow-sm border border-stone-card">
          <h4 className="text-lg font-serif font-bold text-stone-text mb-3">{t('安装方式')}</h4>
          <div className="space-y-3 text-sm">
            <div className="flex items-start gap-3">
              <CheckCircle2 className="w-4 h-4 text-accent-brand mt-0.5 flex-shrink-0" />
              <div>
                <span className="font-medium text-stone-text">{t('方式一 · CLI 安装')}</span>
                <p className="text-stone-muted mt-0.5">{t('运行安装脚本，按提示确认技能目录，自动完成下载与解压。可用环境变量 FASTTOKEN_SKILLS_DIR 指定目录、FASTTOKEN_SKILLS_URL 覆盖下载地址。')}</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <CheckCircle2 className="w-4 h-4 text-accent-brand mt-0.5 flex-shrink-0" />
              <div>
                <span className="font-medium text-stone-text">{t('方式二 · 手动安装')}</span>
                <p className="text-stone-muted mt-0.5">{t('下载 ZIP 后解压，将其中的 fasttoken 文件夹复制到助手的技能目录下即可。')}</p>
              </div>
            </div>
          </div>
        </div>

        {/* 合规声明 */}
        <div className="bg-accent-brand/5 border border-accent-brand/20 rounded-xl p-5">
          <p className="text-sm text-stone-text leading-relaxed">{t('本技能包基于 New API 开源项目衍生，遵循 AGPL-3.0 协议。下载与使用即表示你同意相关许可条款。')}</p>
        </div>
      </div>
    ),
  },
]
}

export function Docs() {
  const { t } = useTranslation()

  const [models, setModels] = useState<ModelInfo[]>([])
  const [pricing, setPricing] = useState<PricingInfo[]>([])
  const [activeSection, setActiveSection] = useState('quick-start')

  useEffect(() => {
    // Fetch available models
    api.get('/api/models', { skipErrorHandler: true } as Record<string, unknown>)
      .then((res) => {
        if (res.data?.success && Array.isArray(res.data?.data)) {
          setModels(res.data.data as ModelInfo[])
        }
      })
      .catch(() => { /* ignore */ })

    // Fetch pricing info
    api.get('/api/pricing', { skipErrorHandler: true } as Record<string, unknown>)
      .then((res) => {
        if (res.data?.success && Array.isArray(res.data?.data)) {
          setPricing(res.data.data as PricingInfo[])
        }
      })
      .catch(() => { /* ignore */ })
  }, [])

  const docSections = buildDocSections(t, models, pricing)
  const currentSection = docSections.find((s) => s.id === activeSection) || docSections[0]

  return (
    <PublicLayout>
      <div className="min-h-screen bg-stone-bg">
        {/* Hero */}
        <div className="bg-card border-b border-stone-card">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-20">
            <div className="text-center">
              <h1 className="text-4xl font-serif font-bold tracking-tight text-stone-text sm:text-5xl lg:text-6xl">{t('用户手册')}</h1>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="flex flex-col lg:flex-row gap-8">
            {/* Sidebar */}
            <div className="lg:w-64 flex-shrink-0">
              <nav className="sticky top-24 space-y-1">
                {docSections.map((section) => (
                  <button
                    key={section.id}
                    type="button"
                    onClick={() => setActiveSection(section.id)}
                    className={`w-full flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition-all duration-200 ${
                      activeSection === section.id
                        ? 'bg-accent-brand/10 text-accent-brand shadow-sm'
                        : 'text-stone-muted hover:bg-card hover:text-stone-text'
                    }`}
                  >
                    <section.icon className={`w-5 h-5 ${activeSection === section.id ? 'text-accent-brand' : 'text-stone-muted/60'}`} />
                    <span>{section.title}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* Main */}
            <div className="flex-1 min-w-0">
              <div className="bg-card rounded-2xl shadow-sm border border-stone-card p-6 sm:p-8
                              hover:shadow-md transition-all duration-300">
                <h2 className="text-2xl font-serif font-bold text-stone-text mb-6 flex items-center gap-3">
                  <currentSection.icon className="w-7 h-7 text-accent-brand" />
                  <span>{currentSection.title}</span>
                </h2>
                <div className="max-w-none">
                  {currentSection.content}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PublicLayout>
  )
}

