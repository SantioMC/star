<script lang="ts">
    import { page } from "$app/state";
    import AuditLogView from "@/components/audit-log-view.svelte";
</script>

<AuditLogView filter={{ application: page.params.app! }} />
