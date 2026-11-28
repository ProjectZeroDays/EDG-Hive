"use client";

import * as React from "react";

import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";

// First-login prompt: while the database is still on the built-in
// hive/hive/hive default credentials (zero-config fresh install), ask the
// operator to rotate the password. The dialog reappears on every login until
// the rotation is done (the server refuses to run it once defaults are gone,
// so this cannot be abused as a generic password changer).
export function DBCredsPrompt() {
  const [open, setOpen] = React.useState(false);
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    let cancelled = false;
    api
      .dbSecurity()
      .then((st) => {
        if (!cancelled && st.default_credentials && st.changeable) setOpen(true);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (password.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }
    if (password !== confirm) {
      toast.error("The two passwords do not match");
      return;
    }
    setSaving(true);
    api
      .dbChangeCredentials(password)
      .then(() => {
        toast.success("Database password changed");
        setOpen(false);
      })
      .catch((err) => toast.error(`Change failed: ${(err as Error).message}`))
      .finally(() => setSaving(false));
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Change the database password</DialogTitle>
          <DialogDescription>
            This install is still using the default database credentials
            (user <span className="font-mono">hive</span>, password{" "}
            <span className="font-mono">hive</span>). Set a new password before
            doing any real work. This prompt reappears on each login until the
            password is changed.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="db-new-password">New database password</Label>
            <Input
              id="db-new-password"
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="At least 8 characters"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="db-confirm-password">Confirm new password</Label>
            <Input
              id="db-confirm-password"
              type="password"
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Later
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? "Changing…" : "Change password"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
