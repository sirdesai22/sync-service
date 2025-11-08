import useSWR from "swr";
import axios from "axios";

const fetcher = (url: string) => axios.get(url).then((res) => res.data);

export default function App() {
  const { data: outbox } = useSWR("http://localhost:8080/api/outbox", fetcher, {
    refreshInterval: 5000,
  });
  console.log(outbox);
  const { data: dlq, mutate } = useSWR("http://localhost:8080/api/dlq", fetcher, {
    refreshInterval: 5000,
  });
  console.log(dlq);
  async function retry(id: string) {
    await axios.get(`http://localhost:8080/api/retry/${id}`);
    mutate(); // refresh DLQ table
  }

  async function addUser() {
    await axios.post("http://localhost:8080/api/add-user");
  }

  async function updateUser() {
    await axios.post("http://localhost:8080/api/update-user");
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen text-gray-800">
      <h1 className="text-2xl font-bold mb-4">Sync Service Dashboard</h1>

      {/* Action Buttons */}
      <div className="flex gap-4 mb-6">
        <button
          onClick={addUser}
          className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
        >
          ➕ Add User
        </button>
        <button
          onClick={updateUser}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          ✏️ Update Random User
        </button>
      </div>
      {/* Outbox Section */}
      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-2">Outbox (latest)</h2>
        <table className="min-w-full bg-white border rounded-lg">
          <thead>
            <tr className="bg-gray-200 text-sm">
              <th className="p-2">ID</th>
              <th className="p-2">Entity</th>
              <th className="p-2">Op</th>
              <th className="p-2">Processed</th>
              <th className="p-2">Created At</th>
            </tr>
          </thead>
          <tbody>
            {Array.isArray(outbox) ? (
              outbox.slice(0, 10).map((o: any) => (
                <tr key={o.ID} className="border-b hover:bg-gray-100 text-sm">
                  <td className="p-2">{o.ID}</td>
                  <td className="p-2">{o.EntityType}</td>
                  <td className="p-2">{o.Op}</td>
                  <td className="p-2">{o.Processed ? "✅" : "❌"}</td>
                  <td className="p-2">
                    {new Date(o.CreatedAt).toLocaleString()}
                  </td>
                </tr>
              ))
            ) : (
              <tr key="no-data">
                <td colSpan={5} className="text-center">
                  No data
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </section>

      {/* DLQ Section */}
      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-2">Dead Letter Queue (DLQ)</h2>
        <table className="min-w-full bg-white border rounded-lg">
          <thead>
            <tr className="bg-gray-200 text-sm">
              <th className="p-2">ID</th>
              <th className="p-2">Entity</th>
              <th className="p-2">Error</th>
              <th className="p-2">Resolved</th>
              <th className="p-2">Action</th>
            </tr>
          </thead>
          <tbody>
            {Array.isArray(dlq) ? (
              dlq.map((d: any) => (
                <tr key={d.id} className="border-b hover:bg-gray-100 text-sm">
                  <td className="p-2">{d.id}</td>
                  <td className="p-2">{d.entity_type}</td>
                  <td className="p-2 text-red-600 truncate max-w-xs">
                    {d.error_msg}
                  </td>
                  <td className="p-2">{d.resolved ? "✅" : "❌"}</td>
                  <td className="p-2">
                    {!d.resolved && (
                      <button
                        onClick={() => retry(d.id)}
                        className="bg-blue-500 text-white px-2 py-1 rounded hover:bg-blue-600"
                      >
                        Retry
                      </button>
                    )}
                  </td>
                </tr>
              ))
            ) : (
              <tr key="no-data">
                <td colSpan={5} className="text-center">
                  No data
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </section>

      {/* Metrics Link */}
      <div className="text-sm text-gray-600">
        <a
          href="http://localhost:2112/metrics"
          target="_blank"
          rel="noreferrer"
        >
          View Prometheus Metrics
        </a>
      </div>
    </div>
  );
}
