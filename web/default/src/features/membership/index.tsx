/*
Copyright (C) 2023-2026 FastToken
*/
import { useTranslation } from 'react-i18next'
import { Link } from '@tanstack/react-router'
import { SectionPageLayout } from '@/components/layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Building2, ArrowRight } from 'lucide-react'
import { MembershipInfoCard } from './components/membership-info-card'
import { InvitationCodeInputCard } from './components/invitation-code-input-card'
import { MembershipUpgradeCard } from './components/membership-upgrade-card'
import { useMembership } from './hooks'

export function Membership() {
  const { t } = useTranslation()
  const { membershipInfo, loading, refresh } = useMembership()

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Membership')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
          <MembershipInfoCard
            membershipInfo={membershipInfo}
            loading={loading}
            onRefresh={refresh}
          />
          <MembershipUpgradeCard
            membershipInfo={membershipInfo}
            loading={loading}
          />
          <InvitationCodeInputCard
            onCodeUsed={refresh}
            disabled={loading}
          />
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2'>
                <Building2 className='h-5 w-5' />
                {t('Enterprise Verification')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className='flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between'>
                <p className='text-sm text-muted-foreground'>
                  {t(
                    'Apply for enterprise verification to get bulk pricing and team management features.',
                  )}
                </p>
                <Button
                  render={
                    <Link to='/enterprise-register'>
                      {t('Apply Now')}
                      <ArrowRight className='ml-1 h-4 w-4' />
                    </Link>
                  }
                  variant='outline'
                  size='sm'
                />
              </div>
            </CardContent>
          </Card>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}