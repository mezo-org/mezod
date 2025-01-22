import { FetchProgressKey } from "./types"

export async function getProgressForKey(
  db: D1Database,
  key: FetchProgressKey,
): Promise<number> {
  const response = await db
    .prepare(
      `
        SELECT page 
        FROM fetch_progress
        WHERE id = ?1
        `,
    )
    .bind(key)
    .first<{ page: number }>()

  return response?.page ?? 0
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
    SET page = ?1, updated_at = datetime('now')
    WHERE id = ?2
    `,
    )
    .bind(newPageCount, key)
    .run()
}
