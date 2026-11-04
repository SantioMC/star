<script lang="ts">
    import { branding } from "@/branding";
    import * as star from "$lib/star";
    import Card from "@/components/card.svelte";
    import EndpointList from "@/components/endpoint-list.svelte";
    import { onMount } from "svelte";

    let discovery = $state<star.DiscoveryDocument | null>(null);
    let loaded = $state(false);

    const identity = $derived(
        discovery
            ? [
                  { label: "Issuer", value: discovery.issuer },
                  { label: "Discovery URL", value: star.getDiscoveryURL() },
                  { label: "JWKS URL", value: discovery.jwks_uri },
              ]
            : [],
    );

    const oauth = $derived(
        discovery
            ? [
                  {
                      label: "Authorization URL",
                      value: discovery.authorization_endpoint,
                  },
                  { label: "Token URL", value: discovery.token_endpoint },
                  { label: "Userinfo URL", value: discovery.userinfo_endpoint },
                  {
                      label: "Revocation URL",
                      value: discovery.revocation_endpoint,
                  },
                  {
                      label: "Introspection URL",
                      value: discovery.introspection_endpoint,
                  },
              ]
            : [],
    );

    onMount(async () => {
        discovery = await star.getDiscovery();
        loaded = true;
    });
</script>

{#if discovery}
    <Card
        title="OpenID Connect"
        icon="id"
        description="Most client libraries only need the discovery URL, the rest can be found from it."
        class="flex flex-col gap-4"
    >
        <EndpointList endpoints={identity} />
    </Card>

    <Card
        title="OAuth 2.0"
        icon="link"
        description="The endpoints this application talks to, shared by every application on {$branding.name}."
        class="flex flex-col gap-4"
    >
        <EndpointList endpoints={oauth} />
    </Card>
{:else if loaded}
    <p class="py-10 text-center text-sm text-muted">
        Failed to load the endpoints for this server.
    </p>
{/if}
