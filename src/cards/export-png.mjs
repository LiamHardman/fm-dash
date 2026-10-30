/** Embed runtime images before export so the complete SVG can be decoded by Canvas. */
export async function cardPngBlob(svgElement, scale = 2) {
  if (!Number.isFinite(scale) || scale <= 0 || scale > 4) throw new RangeError('Export scale must be between 0 and 4.');
  await document.fonts.ready;
  const clone = svgElement.cloneNode(true);
  await Promise.all([...clone.querySelectorAll('image')].map(async image => {
    const source = image.getAttribute('href');
    if (!source || source.startsWith('data:')) return;
    const response = await fetch(new URL(source, document.baseURI));
    if (!response.ok) throw new Error(`Unable to load card image (${response.status}).`);
    const blob = await response.blob();
    const dataUri = await new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result);
      reader.onerror = () => reject(reader.error);
      reader.readAsDataURL(blob);
    });
    image.setAttribute('href', dataUri);
  }));
  const objectUrl = URL.createObjectURL(new Blob([new XMLSerializer().serializeToString(clone)], { type: 'image/svg+xml;charset=utf-8' }));
  try {
    const decoded = new Image();
    decoded.src = objectUrl;
    await decoded.decode();
    const canvas = document.createElement('canvas');
    canvas.width = Math.round(svgElement.viewBox.baseVal.width * scale);
    canvas.height = Math.round(svgElement.viewBox.baseVal.height * scale);
    canvas.getContext('2d').drawImage(decoded, 0, 0, canvas.width, canvas.height);
    return await new Promise((resolve, reject) => canvas.toBlob(blob => blob ? resolve(blob) : reject(new Error('PNG encoding failed.')), 'image/png'));
  } finally {
    URL.revokeObjectURL(objectUrl);
  }
}
