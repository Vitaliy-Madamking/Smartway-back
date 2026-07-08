export const formatPercent = (value: number): string => `${Math.round(value)}%`;

export const formatCoordinates = (latitude: number, longitude: number): string =>
  `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
