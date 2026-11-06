export const opaque = (hex: string) => (hex ?? "#000000").slice(0, 7);

export const alphaOf = (hex: string) => (hex?.length === 9 ? parseInt(hex.slice(7), 16) / 255 : 1);

export const shownAlpha = (hex: string) => Math.round(alphaOf(hex) * 100) / 100;

export function withAlpha(hex: string, alpha: number): string {
  const byte = Math.round(alpha * 255);
  return byte === 255 ? opaque(hex) : opaque(hex) + byte.toString(16).padStart(2, "0");
}
