<script lang="ts">
    import { branding } from "@/branding";
    import * as star from "$lib/star";
    import Card from "@/components/card.svelte";
    import EndpointList from "@/components/endpoint-list.svelte";
    import { getResource } from "@/resource.svelte";

    const resource = getResource<star.SocialProviderConfig>();
    const provider = $derived(resource.current);

    const callback = $derived(
        provider
            ? [
                  {
                      label: "Callback URL",
                      value: star.getProviderCallbackUrl(provider.id),
                  },
              ]
            : [],
    );

    const upstream = $derived.by(() => {
        if (!provider) return [];

        return [
            { label: "Discovery URL", value: provider.oidc_url },
            { label: "Authorization URL", value: provider.authorization_url },
            { label: "Token URL", value: provider.token_url },
            { label: "Userinfo URL", value: provider.userinfo_url },
        ].filter((it): it is { label: string; value: string } => !!it.value);
    });
</script>

<Card
    title="Callback"
    icon="corner-down-left"
    description="Register this URL with the provider so it can redirect users back to {$branding.name}."
    class="flex flex-col gap-4"
>
    <EndpointList endpoints={callback} />
</Card>

{#if upstream.length > 0}
    <Card
        title="Provider Endpoints"
        icon="link"
        description="The OAuth and OpenID Connect endpoints {$branding.name} calls for this provider."
        class="flex flex-col gap-4"
    >
        <EndpointList endpoints={upstream} />
    </Card>
{/if}
