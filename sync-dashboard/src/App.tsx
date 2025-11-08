import axios from "axios";
import { RefreshCcw, Users, TriangleAlert, Database } from "lucide-react";
import useSWR from "swr";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const fetcher = (url: string) => axios.get(url).then((res) => res.data);

export default function App() {
  const {
    data: outbox,
    mutate: refreshOutbox,
    isLoading: outboxLoading,
  } = useSWR("http://localhost:8080/api/outbox", fetcher, {
    refreshInterval: 5000,
  });
  const {
    data: dlq,
    mutate: refreshDlq,
    isLoading: dlqLoading,
  } = useSWR("http://localhost:8080/api/dlq", fetcher, {
    refreshInterval: 5000,
  });

  const safeOutbox: any[] = Array.isArray(outbox) ? outbox : [];
  const safeDlq: any[] = Array.isArray(dlq) ? dlq : [];

  const processedCount = safeOutbox.filter(
    (item) => item.Processed ?? item.processed,
  ).length;
  const pendingCount = safeOutbox.length - processedCount;
  const activeDlq = safeDlq.filter((item) => !(item.Resolved ?? item.resolved))
    .length;

  async function retry(id: string) {
    await axios.get(`http://localhost:8080/api/retry/${id}`);
    refreshDlq();
    refreshOutbox();
  }

  async function addUser() {
    await axios.post("http://localhost:8080/api/add-user");
    refreshOutbox();
  }

  async function updateUser() {
    await axios.post("http://localhost:8080/api/update-user");
    refreshOutbox();
  }

  return (
    <div className="min-h-screen bg-slate-950 py-12 text-slate-50">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-6">
        <header className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm uppercase tracking-wide text-slate-400">
              Sync Service
            </p>
            <h1 className="text-3xl font-semibold tracking-tight">
              Operational Dashboard
            </h1>
            <p className="text-sm text-slate-400">
              Monitor the outbox pipeline, DLQ recovery, and trigger manual actions.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button variant="secondary" onClick={() => refreshOutbox()}>
              <RefreshCcw className="mr-2 h-4 w-4" />
              Refresh Data
            </Button>
            <Button className="hover:border-white border-white" onClick={addUser}>
              <Users className="mr-2 h-4 w-4" />
              Add Sample User
            </Button>
            <Button variant="outline" className="text-black" onClick={updateUser}>
              <Database className="mr-2 h-4 w-4" />
              Update Random User
            </Button>
          </div>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <Card className="bg-slate-900/40 backdrop-blur">
            <CardHeader className="pb-2">
              <CardDescription>Outbox events</CardDescription>
              <CardTitle className="text-3xl text-white">{safeOutbox.length}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-slate-400">
                {outboxLoading ? "Loading latest batch…" : "Latest 100 records shown"}
              </p>
            </CardContent>
          </Card>
          <Card className="bg-slate-900/40 backdrop-blur">
            <CardHeader className="pb-2">
              <CardDescription>Pending sync</CardDescription>
              <CardTitle className="text-3xl text-white">{pendingCount}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-slate-400">
                {processedCount} already processed
              </p>
            </CardContent>
          </Card>
          <Card className="bg-slate-900/40 backdrop-blur">
            <CardHeader className="pb-2 flex-row items-center justify-between gap-2">
              <div>
                <CardDescription>DLQ alerts</CardDescription>
                <CardTitle className="text-3xl text-white">{activeDlq}</CardTitle>
              </div>
              <TriangleAlert className="h-6 w-6 text-amber-400" />
            </CardHeader>
            <CardContent>
              <p className="text-sm text-slate-400">
                {dlqLoading ? "Checking DLQ…" : "Unresolved items awaiting retry"}
              </p>
            </CardContent>
          </Card>
        </section>

        <section className="grid gap-6 lg:grid-cols-2">
          <Card className="bg-slate-900/40 backdrop-blur">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Latest Outbox Events</CardTitle>
                  <CardDescription>
                    Observing newest items by descending ID.
                  </CardDescription>
                </div>
                <Badge variant="secondary">
                  Showing {Math.min(10, safeOutbox.length)} of {safeOutbox.length}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="overflow-hidden rounded-lg border border-slate-800">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-20">ID</TableHead>
                      <TableHead>Entity</TableHead>
                      <TableHead>Operation</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="w-40">Created</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {safeOutbox.slice(0, 10).map((item) => {
                      const id = item.ID ?? item.id;
                      const entity = item.EntityType ?? item.entity_type;
                      const op = item.Op ?? item.op;
                      const processed = item.Processed ?? item.processed;
                      const created = item.CreatedAt ?? item.created_at;
                      return (
                        <TableRow key={id}>
                          <TableCell className="font-medium">{id}</TableCell>
                          <TableCell className="capitalize">{entity}</TableCell>
                          <TableCell>
                            <Badge variant="outline">{op}</Badge>
                          </TableCell>
                          <TableCell>
                            {processed ? (
                              <Badge variant="secondary">Processed</Badge>
                            ) : (
                              <Badge variant="destructive">Pending</Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-xs text-slate-400">
                            {created ? new Date(created).toLocaleString() : "—"}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                    {safeOutbox.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={5} className="text-center text-slate-400">
                          No outbox events yet.
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-slate-900/40 backdrop-blur">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Dead Letter Queue</CardTitle>
                  <CardDescription>
                    Items that require manual intervention or retry.
                  </CardDescription>
                </div>
                <Badge variant="destructive">{activeDlq} unresolved</Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="overflow-hidden rounded-lg border border-slate-800">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">ID</TableHead>
                      <TableHead>Entity</TableHead>
                      <TableHead>Error</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="w-32 text-right">Action</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {safeDlq.slice(0, 10).map((item) => {
                      const id = item.id ?? item.ID;
                      const entity = item.entity_type ?? item.EntityType;
                      const error = item.error_msg ?? item.ErrorMsg;
                      const resolved = item.resolved ?? item.Resolved;
                      return (
                        <TableRow key={id}>
                          <TableCell className="font-medium">{id}</TableCell>
                          <TableCell className="capitalize">{entity}</TableCell>
                          <TableCell className="max-w-xs truncate text-xs text-red-300">
                            {error}
                          </TableCell>
                          <TableCell>
                            {resolved ? (
                              <Badge variant="secondary">Resolved</Badge>
                            ) : (
                              <Badge variant="destructive">Open</Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="outline"
                              size="sm"
                              disabled={resolved}
                              onClick={() => retry(id)}
                            >
                              Retry
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                    {safeDlq.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={5} className="text-center text-slate-400">
                          DLQ is empty 🎉
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </section>

        <footer className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-800 bg-slate-900/40 px-6 py-4 backdrop-blur">
          <div>
            <p className="text-sm font-medium text-slate-200">
              Observability endpoints
            </p>
            <p className="text-xs text-slate-400">
              Prometheus metrics exposed by the sync service.
            </p>
          </div>
          <Button variant="outline" asChild>
            <a href="http://localhost:2112/metrics" target="_blank" rel="noreferrer">
              View Prometheus Metrics
            </a>
          </Button>
        </footer>
      </div>
    </div>
  );
}
