export default function StatsCard({ stats }) {
  return (
    <div className="mb-6 p-4 bg-slate-800 rounded-lg inline-block">
      <span className="text-4xl font-bold text-green-400">
        {stats.msgPerSec}
      </span>

      <span className="ml-2 text-slate-400">msgs/sec</span>

      <div className="text-xs text-slate-500 mt-1">
        total: {stats.messagesReceived} | duplicate skipped:{" "}
        {stats.duplicatesSkipped} | OutOfOrder skipped:{" "}
        {stats.outOfOrderSkipped}
      </div>
    </div>
  );
}
