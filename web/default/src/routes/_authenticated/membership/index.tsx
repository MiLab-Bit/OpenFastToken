import { createFileRoute } from '@tanstack/react-router'
import { Membership } from '@/features/membership'

export const Route = createFileRoute('/_authenticated/membership/')({
  component: Membership,
})