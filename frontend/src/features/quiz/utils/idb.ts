import type { BatchAnswerItem } from '@/types/generated';

const DB_NAME = 'penlight_offline_db';
const DB_VERSION = 1;
const STORE_ANSWERS = 'answer_outbox';

function getIndexedDB(): IDBFactory | null {
  if (typeof window === 'undefined' || !window.indexedDB) {
    return null;
  }
  return window.indexedDB;
}

function openDB(): Promise<IDBDatabase> {
  const idb = getIndexedDB();
  if (!idb) {
    return Promise.reject(
      new Error('IndexedDB is not available in this environment'),
    );
  }

  return new Promise((resolve, reject) => {
    const request = idb.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;
      if (!db.objectStoreNames.contains(STORE_ANSWERS)) {
        db.createObjectStore(STORE_ANSWERS, { keyPath: 'id' });
      }
    };

    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

/**
 * Enqueue an answer record into IndexedDB outbox.
 */
export async function enqueueAnswer(item: BatchAnswerItem): Promise<void> {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_ANSWERS, 'readwrite');
      const store = tx.objectStore(STORE_ANSWERS);
      const req = store.put(item);

      req.onsuccess = () => resolve();
      req.onerror = () => reject(req.error);
    });
  } catch (err) {
    console.warn('Failed to enqueue answer to IndexedDB:', err);
  }
}

/**
 * Retrieve all pending answer records from IndexedDB outbox.
 */
export async function getPendingAnswers(): Promise<BatchAnswerItem[]> {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_ANSWERS, 'readonly');
      const store = tx.objectStore(STORE_ANSWERS);
      const req = store.getAll();

      req.onsuccess = () => resolve(req.result as BatchAnswerItem[]);
      req.onerror = () => reject(req.error);
    });
  } catch (err) {
    console.warn('Failed to get pending answers from IndexedDB:', err);
    return [];
  }
}

/**
 * Remove successfully synchronized answers from IndexedDB outbox.
 */
export async function removePendingAnswers(ids: string[]): Promise<void> {
  if (ids.length === 0) return;
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_ANSWERS, 'readwrite');
      const store = tx.objectStore(STORE_ANSWERS);

      for (const id of ids) {
        store.delete(id);
      }

      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error);
    });
  } catch (err) {
    console.warn('Failed to remove pending answers from IndexedDB:', err);
  }
}
