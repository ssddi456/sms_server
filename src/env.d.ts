export {};

declare global {
  interface Window {
    __SMS_UI_CONFIG__?: {
      apkFileName?: string;
    };
  }
}
