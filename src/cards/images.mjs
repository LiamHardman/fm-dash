const imageLoads = new Map();

/** Browser-only readiness check, cached by URL. Remote export still requires CORS. */
export function waitForCardImage(source) {
  if (!source) return Promise.resolve('');
  if (!imageLoads.has(source)) imageLoads.set(source, new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = async () => {
      try { await image.decode(); resolve(source); } catch (error) { reject(error); }
    };
    image.onerror = () => reject(new Error('A card image could not be loaded.'));
    image.src = source;
  }));
  return imageLoads.get(source);
}

/** Art is required; optional images fall back to the renderer's silhouette/code tiles. */
export async function prepareCardImages({ artwork, portrait, nation, club }) {
  const [art, face, flag, badge] = await Promise.all([
    waitForCardImage(artwork),
    waitForCardImage(portrait).catch(() => ''),
    waitForCardImage(nation).catch(() => ''),
    waitForCardImage(club).catch(() => ''),
  ]);
  return { artwork: art, portrait: face, nation: flag, club: badge };
}
