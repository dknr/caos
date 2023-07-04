type Attrs = Record<string, string>;
const attrify = (attrs: Attrs): string =>
  Object.entries(attrs).map(([key, value]) => 
    ` ${key}="${value.replaceAll('"', '&quot;')}"`
  ).join('');

type ElParams = string[] | [Attrs, ...string[]];
const unpack = (params: ElParams): [Attrs, string[]] => {
  const [attrs, inner] = typeof params[0] === 'object'
    ? [params[0], params.slice(1) as string[]]
    : [{}, params as string[]];
    
  return [attrs, inner];
}

export const el = (el: string, sep = '\n') => (...params: ElParams): string => {
  const [_attrs, inner] = unpack(params);
  const attrs = attrify(_attrs);

  if (inner) {
    if (inner.length > 1) {
      return `<${el}${attrs}>${sep}${inner.join(sep)}${sep}</${el}>`
    } else {
      return `<${el}${attrs}>${inner[0]}</${el}>`
    }
  } else {
    return `<${el}${attrs}/>`;
  }
}


const kebab = (input: string) =>{
  return input.split(/(?=[A-Z])/).map((part) => part.toLowerCase()).join('-');
}

type CssProperties = Record<string, string>;
const stylify = (properties: CssProperties): string => {
  return Object.entries(properties).map(
    ([property, value]) => `${kebab(property)}: ${value};`
  ).join('\n');
}
const css = (styles: Record<string, CssProperties>): string => {
  return Object.entries(styles).map(
    ([selector, properties]) => `${selector} {\n${stylify(properties)}\n}`
  ).join('\n');
}

export const style = (styles: Record<string, CssProperties>) => {
  return el('style')(css(styles));
}



export const html = el('html');
export const head = el('head');
export const title = (title: string) => el('title')(title);
export const body = el('body');
export const div = el('div');
export const span = el('span');
export const a = el('a');
export const h = (n: 1 | 2 | 3 | 4 | 5) => el(`h${n}`);
export const p = el('p');
export const pre = el('pre', '');

export const page = (name: string, ...inner: string[]) => html(
  head(
    title(`caos - ${name}`),
    style({
      body: {
        backgroundColor: 'black',
        color: 'white',
      },
      '.path': {
        fontSize: '16px',
        fontFamily: 'monospace',
      }
    })
  ),
  body(
    ...inner
  )
);
