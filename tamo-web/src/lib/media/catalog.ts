export type CatalogAsset = { id: string; src: string; alt: string; orientation: 'portrait' | 'landscape' | 'square'; category: string; kind?: 'image'|'video'|'illustration' };
export const catalogAssets: CatalogAsset[] = [
  { id: 'football-01', src: '/images/football-action.png', alt: 'Football players competing for the ball on a green pitch', orientation: 'portrait', category: 'sport' },
  { id: 'classroom-01', src: '/images/classroom-child.png', alt: 'A young pupil writing in a classroom', orientation: 'portrait', category: 'education' },
  { id: 'city-bus-01', src: '/images/city-bus.png', alt: 'A commuter boarding a city bus', orientation: 'portrait', category: 'travel' },
  { id: 'portrait-01', src: '/images/young-woman-portrait.png', alt: 'Portrait of a young woman wearing a hijab', orientation: 'portrait', category: 'people' },
];
export const galleryAssets = [...catalogAssets, ...catalogAssets].map((asset, index) => ({ ...asset, id: `${asset.id}-${index + 1}` }));
