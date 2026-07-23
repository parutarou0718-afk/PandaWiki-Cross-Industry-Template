import { DomainPublicEditionConfig } from '@/request/types';

const knownEditions = new Set(['common', 'research', 'legal', 'finance']);

/** Returns only a valid public edition. Missing or invalid data keeps legacy UI behavior. */
export const getPublicEdition = (
  edition?: DomainPublicEditionConfig,
): DomainPublicEditionConfig | undefined =>
  edition && edition.edition_id && knownEditions.has(edition.edition_id)
    ? edition
    : undefined;

export const editionTerm = (
  edition: DomainPublicEditionConfig | undefined,
  key: string,
  fallback: string,
) => edition?.terminology?.[key] || fallback;
