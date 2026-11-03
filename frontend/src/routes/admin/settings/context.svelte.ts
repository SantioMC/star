import * as star from "$lib/star";
import { getContext, setContext } from "svelte";
import { toast } from "svelte-sonner";

const key = Symbol("server-settings");

export class SettingsContext {
  current = $state<star.ServerSettings | null>(null);

  load = async () => {
    const result = await star.getServerSettings();
    if (!result) {
      toast.error("Failed to load the server settings");
      return;
    }

    this.current = result;
  };

  update = async (changes: star.UpdateServerSettingsRequest) => {
    const result = await star.updateServerSettings(changes);
    if ("error" in result) {
      toast.error(result.message);
      return false;
    }

    this.current = result;
    toast.success("Successfully updated the server settings");
    return true;
  };
}

export const setSettingsContext = (context: SettingsContext) =>
  setContext(key, context);

export const getSettingsContext = () => getContext<SettingsContext>(key);
