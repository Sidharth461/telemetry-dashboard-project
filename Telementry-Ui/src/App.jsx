import React, { useEffect, useState } from "react";
import Header from "./components/Header";
import DeviceCard from "./components/DeviceCard";
import StatsCard from "./components/StatsCard";
const App = () => {
  const [devices, setDevices] = useState([]);
  const [stats, setStats] = useState({
    msgPerSec: 0,
    messagesReceived: 0,
    duplicatesSkipped: 0,
    outOfOrderSkipped: 0,
  });

  useEffect(() => {
    const eventSource = new EventSource("http://localhost:8080/api/stream");

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setDevices(data.devices);
      setStats(data.stats);
    };

    return () => eventSource.close();
  }, []);

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 p-6">
      <Header />

      <StatsCard stats={stats} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {devices.map((device) => (
          <DeviceCard key={device.deviceId} device={device} />
        ))}
      </div>
    </div>
  );
};

export default App;
