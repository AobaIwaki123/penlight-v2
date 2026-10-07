/**
 * Generates an RFC 9562 compliant UUID v7 identifier.
 */
function generateUUIDv7(): string {
  const now = Date.now();
  const bytes = new Uint8Array(16);
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(bytes);
  } else {
    for (let i = 0; i < 16; i++) {
      bytes[i] = Math.floor(Math.random() * 256);
    }
  }

  // 48-bit timestamp Big-Endian
  bytes[0] = Math.floor(now / 0x10000000000) & 0xff;
  bytes[1] = Math.floor(now / 0x100000000) & 0xff;
  bytes[2] = Math.floor(now / 0x1000000) & 0xff;
  bytes[3] = Math.floor(now / 0x10000) & 0xff;
  bytes[4] = Math.floor(now / 0x100) & 0xff;
  bytes[5] = now & 0xff;

  // Version 7 (0b0111_xxxx)
  bytes[6] = 0x70 | (bytes[6] & 0x0f);
  // Variant RFC 4122 (0b10xx_xxxx)
  bytes[8] = 0x80 | (bytes[8] & 0x3f);

  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join(
    '',
  );
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

/**
 * Generate a TypeID for answer logs (ans_<uuidv7>)
 */
export function generateAnswerID(): string {
  return `ans_${generateUUIDv7()}`;
}

/**
 * Generate a TypeID for quiz questions (quiz_<uuidv7>)
 */
export function generateQuizQuestionID(): string {
  return `quiz_${generateUUIDv7()}`;
}
