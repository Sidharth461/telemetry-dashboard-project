export default function DeviceCard({ device }) {
  const secondsSince = Math.floor(
    (Date.now() - new Date(device.lastMessageAt).getTime()) / 1000,
  );
  return (
    <div className="p-4 rounded-lg border bg-slate-800 border-slate-600">
      <div className="flex justify-between">
        <span className="font-bold">{device.deviceId}</span>
        <span>{device.isOnline ? device.state : "OFFLINE"}</span>
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
          <span>{secondsSince}s ago</span>
        </div>
      </div>
    </div>
  );
}
