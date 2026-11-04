import { getContext, setContext } from "svelte";

const key = Symbol("admin-resource");

export class Resource<T> {
  current = $state<T | null>(null);
}

export const setResource = <T>(resource: Resource<T>) =>
  setContext(key, resource);

export const getResource = <T>() => getContext<Resource<T>>(key);
