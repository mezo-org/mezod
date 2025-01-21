import { FetchProgressKey } from "./types"

export async function getProgressForKey(
  db: D1Database,
  key: FetchProgressKey,
): Promise<number> {
  const response = await db
    .prepare(
      `
        SELECT page_count 
        FROM fetch_progress
        WHERE id = ?1
        `,
    )
    .bind(key)
    .first<{ page_count: number }>()

  return response?.page_count ?? 0
}

export async function updateProgressForKey(
  db: D1Database,
  key: FetchProgressKey,
  newPageCount: number,
) {
  await db
    .prepare(
      `
    UPDATE fetch_progress
    SET page_count = ?1, updated_at = datetime('now')
    WHERE id = ?2
    `,
    )
    .bind(newPageCount, key)
    .run()
}
