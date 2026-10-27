import { HistoryPanel } from "./HistoryPanel";

interface FeaturesModuleProps {
  reportKey: string;
  onRequireSettings: () => void;
}

export function FeaturesModule({ reportKey, onRequireSettings }: FeaturesModuleProps) {
  return (
    <section className="module-grid">
      <HistoryPanel reportKey={reportKey} onRequireSettings={onRequireSettings} />
    </section>
  );
}
