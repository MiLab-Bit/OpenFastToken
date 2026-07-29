import { createFileRoute } from '@tanstack/react-router'
import { RechargeManagement } from '@/features/recharge-admin'

export const Route = createFileRoute('/_authenticated/recharge-admin/')({
  component: RechargeManagement,
})
