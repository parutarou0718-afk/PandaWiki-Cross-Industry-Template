import {
  FeatureStatus,
  VersionInfoMap,
  VersionInfo,
  getFeatureValue,
} from '@/constant/version';
import { ConstsLicenseEdition } from '@/request/types';

export const useFeatureValue = <K extends keyof VersionInfo['features']>(
  key: K,
): VersionInfo['features'][K] => {
  return getFeatureValue(ConstsLicenseEdition.LicenseEditionEnterprise, key);
};

export const useFeatureValueSupported = (
  key: keyof VersionInfo['features'],
) => {
  return (
    getFeatureValue(ConstsLicenseEdition.LicenseEditionEnterprise, key) ===
      FeatureStatus.SUPPORTED ||
    getFeatureValue(ConstsLicenseEdition.LicenseEditionEnterprise, key) ===
      FeatureStatus.ADVANCED
  );
};

export const useVersionInfo = () => {
  return VersionInfoMap[ConstsLicenseEdition.LicenseEditionEnterprise];
};
