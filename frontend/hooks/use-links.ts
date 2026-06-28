"use client"

import {
  useMutation,
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query"

import { links } from "@/lib/api"
import type {
  CreateLinkInput,
  ListLinksParams,
  ListLinksResponse,
  UpdateLinkInput,
} from "@/types/api"

// Query-key factory for link data. Centralising keys here keeps invalidation
// consistent across the dashboard (CLAUDE.md).
export const linkKeys = {
  all: ["links"] as const,
  lists: () => [...linkKeys.all, "list"] as const,
  list: (params: ListLinksParams) => [...linkKeys.lists(), params] as const,
  detail: (slug: string) => [...linkKeys.all, "detail", slug] as const,
}

/** Paginated, searchable, sortable list of the current user's links. */
export function useLinks(params: ListLinksParams) {
  return useQuery<ListLinksResponse>({
    queryKey: linkKeys.list(params),
    queryFn: () => links.list(params),
    placeholderData: keepPreviousData,
  })
}

/** Creates a link and refreshes every list query. */
export function useCreateLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateLinkInput) => links.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.lists() })
    },
  })
}

/** Updates a link by slug and refreshes lists. */
export function useUpdateLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ slug, input }: { slug: string; input: UpdateLinkInput }) =>
      links.update(slug, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.lists() })
    },
  })
}

/** Deletes a link by slug and refreshes lists. */
export function useDeleteLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (slug: string) => links.remove(slug),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: linkKeys.lists() })
    },
  })
}
