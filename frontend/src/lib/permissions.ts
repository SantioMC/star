export interface PermissionAction {
  name: string;
  action: string;
}

export interface PermissionResource {
  resource: string;
  actions: PermissionAction[];
}

export interface PermissionNamespace {
  namespace: string;
  count: number;
  resources: PermissionResource[];
}

const permissionNamespace = (name: string): string =>
  name.split(":")[0];

const splitPermission = (name: string): PermissionAction & { resource: string } => {
  const separator = name.lastIndexOf(":");
  if (separator === -1) return { name, resource: name, action: name };

  return {
    name,
    resource: name.slice(0, separator),
    action: name.slice(separator + 1),
  };
};

export const groupByNamespace = <T>(
  items: T[],
  nameOf: (item: T) => string,
): [string, T[]][] => {
  const groups = new Map<string, T[]>();
  const sorted = [...items].sort((a, b) => nameOf(a).localeCompare(nameOf(b)));

  for (const item of sorted) {
    const namespace = permissionNamespace(nameOf(item));
    const group = groups.get(namespace) ?? [];

    groups.set(namespace, [...group, item]);
  }

  return [...groups];
};

export const groupPermissions = (names: string[]): PermissionNamespace[] =>
  groupByNamespace(names, (name) => name).map(([namespace, members]) => {
    const resources = new Map<string, PermissionAction[]>();

    for (const name of members) {
      const { resource, action } = splitPermission(name);
      const actions = resources.get(resource) ?? [];

      resources.set(resource, [...actions, { name, action }]);
    }

    return {
      namespace,
      count: members.length,
      resources: [...resources]
        .sort(([a], [b]) => Number(b === namespace) - Number(a === namespace))
        .map(([resource, actions]) => ({ resource, actions })),
    };
  });
