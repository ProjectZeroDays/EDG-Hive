"use client";

import Link from "next/link";

import { useShallow } from "zustand/react/shallow";

import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader } from "@/components/ui/sidebar";
import { useCurrentUser } from "@/hooks/use-current-user";
import { sidebarItems } from "@/navigation/sidebar/sidebar-items";
import { usePreferencesStore } from "@/stores/preferences/preferences-provider";

import { NavMain } from "./nav-main";
import { NavUser } from "./nav-user";
import { SidebarSupportCard } from "./sidebar-support-card";


export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const currentUser = useCurrentUser();
  const { sidebarVariant, sidebarCollapsible, isSynced } = usePreferencesStore(
    useShallow((s) => ({
      sidebarVariant: s.sidebarVariant,
      sidebarCollapsible: s.sidebarCollapsible,
      isSynced: s.isSynced,
    })),
  );

  const variant = isSynced ? sidebarVariant : props.variant;
  const collapsible = isSynced ? sidebarCollapsible : props.collapsible;

  return (
    <Sidebar {...props} variant={variant} collapsible={collapsible}>
      <SidebarHeader>
        <Link prefetch={false} href="/function/tasks" className="flex items-center gap-4 rounded-md p-2">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src="/hive-wordmark.png"
            alt="Hive"
            width={900}
            height={220}
            className="hive-wordmark h-[37px] w-auto shrink-0"
          />
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src="/cia-seal.png"
            alt=""
            width={512}
            height={512}
            className="size-[41px] shrink-0 rounded-full ring-2 ring-white"
          />
        </Link>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={sidebarItems} />
        {/* <NavDocuments items={data.documents} /> */}
        {/* <NavSecondary items={data.navSecondary} className="mt-auto" /> */}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={currentUser} />
      </SidebarFooter>
    </Sidebar>
  );
}
