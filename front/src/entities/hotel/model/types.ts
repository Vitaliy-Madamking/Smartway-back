export type Coordinates = {
  latitude: number;
  longitude: number;
};

export type HotelRecord = {
  id: string;
  providerId: string;
  externalId: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  coordinates?: Coordinates;
};
