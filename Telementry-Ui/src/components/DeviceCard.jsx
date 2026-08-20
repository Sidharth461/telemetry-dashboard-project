export default function DeviceCard({ device }) {
  const secondsSince = Math.floor(
    (Date.now() - new Date(device.lastMessageAt).getTime()) / 1000,
  );

  const online = device.isOnline;

  return (
    <div
      className={`p-4 rounded-lg border ${
        online
          ? "bg-slate-800 border-green-500"
          : "bg-slate-800 border-red-500 opacity-70"
      }`}
    >
      <div className="flex justify-between items-center">
        <span className="font-bold">{device.deviceId}</span>
        <span
          className={`px-2 py-0.5 rounded text-xs font-bold ${
            online
              ? device.state === "MOVING"
                ? "bg-green-500/20 text-green-400"
                : "bg-yellow-500/20 text-yellow-400"
              : "bg-red-500/20 text-red-400"
          }`}
        >
          {online ? device.state : "OFFLINE"}
        </span>
      </div>

      <div className="mt-3 space-y-1">
        <div className="flex justify-between">
          <span>Speed</span>
          <span>{device.speedKph.toFixed(1)} km/h</span>
        </div>

        <div className="flex justify-between">
          <span>Battery</span>
          <span>{device.batteryPct.toFixed(0)}%</span>
        </div>

        <div className="flex justify-between">
          <span>Moving</span>
          <span>{device.movingPercent.toFixed(0)}%</span>
        </div>

        <div className="flex justify-between">
          <span>Distance</span>
          <span>{device.distanceMeters.toFixed(0)} m</span>
        </div>

        <div className="flex justify-between">
          <span>Last msg</span>
          <span className={online ? "text-green-400" : "text-red-400"}>
            {secondsSince}s ago
          </span>
        </div>
      </div>
    </div>
  );
}
