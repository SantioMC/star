<script lang="ts">
    import { page } from "$app/state";
    import AuditLogView from "@/components/audit-log-view.svelte";
</script>

<AuditLogView filter={{ user: page.params.user! }} />
