type FilterFlags<Base, Condition> = {
  [Key in keyof Base]: Base[Key] extends Condition ? Key : never
};
type AllowedNames<Base, Condition> = FilterFlags<Base, Condition>[keyof Base];
export type SubType<Base, Condition> = Pick<
  Base,
  AllowedNames<Base, Condition>
>;

export const getPatchOpsFromObj = <T>(
  allowedKeyForReplace: (keyof T)[],
  payload: T
) => {
  return (Object.keys(payload) as (keyof T)[])
    .filter(key => allowedKeyForReplace.includes(key))
    .reduce((prev, key) => {
      return [
        ...prev,
        {
          op: 'replace',
          path: `/${key}`,
          value: payload[key] || ''
        }
      ];
    }, []);
};

const propReducer = (key, payload) => (prevProp, keyProp) => {
  return [
    ...prevProp,
    {
      op: payload[keyProp] ? 'add' : 'remove',
      path: `/${key}/${keyProp}`
    }
  ];
};

export const getPatchAddRemoveOpsFromObj = <T>(
  allowedKeyForAddRemove: (keyof T)[],
  payload: T
) => {
  return (Object.keys(payload) as (keyof T)[])
    .filter(key => allowedKeyForAddRemove.includes(key))
    .reduce((prev, key) => {
      return [
        ...prev,
        ...Object.keys(payload[key]).reduce(propReducer(key, payload[key]), [])
      ];
    }, []);
};
