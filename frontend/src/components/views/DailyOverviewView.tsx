import { useCallback, useEffect, useLayoutEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as Accordion from '@radix-ui/react-accordion'
import styled from 'styled-components'
import { useGTLocalStorage } from '../../hooks'
import { Border, Colors, Spacing, Typography } from '../../styles'
import { icons } from '../../styles/images'
import { TPullRequest, TTask } from '../../utils/types'
import Flex from '../atoms/Flex'
import Spinner from '../atoms/Spinner'
import GTButton from '../atoms/buttons/GTButton'
import EmptyDetails from '../details/EmptyDetails'
import PullRequestDetails from '../details/PullRequestDetails'
import TaskDetails from '../details/TaskDetails'
import { SectionHeader } from '../molecules/Header'
import OverviewListsModal from '../overview/OverviewListsModal'
import useOverviewLists from '../overview/useOverviewLists'
import OverviewAccordionItem from '../radix/OverviewAccordionItem'
import ScrollableListTemplate from '../templates/ScrollableListTemplate'

const ActionsContainer = styled.div`
    background-color: ${Colors.background.medium};
    padding: ${Spacing._8} ${Spacing._12};
    border-radius: ${Border.radius.mini};
    display: flex;
    gap: ${Spacing._24};
    margin-bottom: ${Spacing._16};
`
const BannerButton = styled(GTButton)`
    ${Typography.label};
`
const AccordionRoot = styled(Accordion.Root)`
    > * > h3 {
        all: unset;
    }
    > div {
        margin-bottom: ${Spacing._4};
    }
`
const RightActions = styled.div`
    margin-left: auto;
    display: flex;
`

const DailyOverviewView = () => {
    const { lists, isLoading } = useOverviewLists()
    const [openListIds, setOpenListIds] = useState<string[]>([])
    const [isEditListsModalOpen, setIsEditListsModalOpen] = useState(false)
    const [editListTabIndex, setEditListTabIndex] = useState(0) // 0 - add, 1 - reorder
    const { overviewViewId, overviewItemId, subtaskId } = useParams()
    const navigate = useNavigate()
    const [overviewAutomaticEmptySort] = useGTLocalStorage('overviewAutomaticEmptySort', false, true)

    if (overviewAutomaticEmptySort) {
        lists.sort((a, b) => {
            if (a.view_items.length === 0 && b.view_items.length > 0) return 1
            if (a.view_items.length > 0 && b.view_items.length === 0) return -1
            return 0
        })
    }

    const selectFirstItem = () => {
        const firstNonEmptyView = lists?.find((list) => list.view_items.length > 0)
        if (firstNonEmptyView) {
            navigate(`/daily-overview/${firstNonEmptyView.id}/${firstNonEmptyView.view_items[0].id}`, { replace: true })
        }
    }

    const detailsView = useMemo(() => {
        if (!lists?.length) return <EmptyDetails icon={icons.list} text="You have no views" />
        for (const list of lists) {
            if (list.id !== overviewViewId) continue
            for (const item of list.view_items) {
                if (item.id !== overviewItemId) continue
                if (list.type === 'github') return <PullRequestDetails pullRequest={item as TPullRequest} />

                const subtask = item?.sub_tasks?.find((subtask) => subtask.id === subtaskId)
                const detailsLink = subtask
                    ? `/daily-overview/${list.id}/${item.id}/${subtask.id}`
                    : `/daily-overview/${list.id}/${item.id}`
                return <TaskDetails task={item as TTask} subtask={subtask} link={detailsLink} />
            }
        }
        return null
    }, [lists, overviewItemId, overviewViewId, subtaskId])

    useLayoutEffect(() => {
        const firstNonEmptyList = lists?.find((list) => list.view_items.length > 0)
        if (firstNonEmptyList) setOpenListIds([firstNonEmptyList.id])
    }, [])

    const removeListFromOpenListIds = (id: string) => setOpenListIds(openListIds.filter((value) => value !== id))

    useEffect(() => {
        if (!isLoading && (!overviewViewId || !overviewItemId || !detailsView)) {
            selectFirstItem()
        }
        // check that selected item is in list of views
        for (const list of lists) {
            if (list.id === overviewViewId) {
                for (const item of list.view_items) {
                    if (item.id === overviewItemId) {
                        return
                    }
                }
            }
        }
        selectFirstItem()
    }, [isLoading, overviewViewId, overviewItemId, lists, detailsView])

    const collapseAll = () => setOpenListIds([])
    const expandAll = useCallback(() => setOpenListIds(lists.map((list) => list.id)), [lists])

    if (isLoading) return <Spinner />
    return (
        <>
            <Flex>
                <ScrollableListTemplate>
                    <SectionHeader sectionName="Daily Overview" />
                    <ActionsContainer>
                        <BannerButton
                            styleType="simple"
                            size="small"
                            onClick={() => {
                                setEditListTabIndex(1)
                                setIsEditListsModalOpen(true)
                            }}
                            icon={icons.bolt}
                            iconColor="gray"
                            value={
                                <span>
                                    Smart Prioritize<sup>AI</sup>
                                </span>
                            }
                        />
                        <RightActions>
                            <BannerButton
                                styleType="simple"
                                size="small"
                                onClick={collapseAll}
                                icon={icons.squareMinus}
                                iconColor="gray"
                                value="Collapse all"
                            />
                            <BannerButton
                                styleType="simple"
                                size="small"
                                onClick={expandAll}
                                icon={icons.squarePlus}
                                iconColor="gray"
                                value="Expand all"
                            />
                            <BannerButton
                                styleType="simple"
                                size="small"
                                onClick={() => {
                                    setEditListTabIndex(0)
                                    setIsEditListsModalOpen(true)
                                }}
                                icon={icons.squarePlus}
                                iconColor="gray"
                                value="Edit lists"
                            />
                        </RightActions>
                    </ActionsContainer>
                    <AccordionRoot type="multiple" value={openListIds} onValueChange={setOpenListIds}>
                        {lists.map((list) => (
                            <OverviewAccordionItem
                                key={list.id}
                                list={list}
                                closeAccordion={() => removeListFromOpenListIds(list.id)}
                            />
                        ))}
                    </AccordionRoot>
                </ScrollableListTemplate>
            </Flex>
            {detailsView}
            <OverviewListsModal
                isOpen={isEditListsModalOpen}
                setisOpen={setIsEditListsModalOpen}
                defaultTabIndex={editListTabIndex}
            />
        </>
    )
}

export default DailyOverviewView
