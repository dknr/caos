const time = () => {
  const now = new Date();
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2,'0')}:${now.getSeconds().toString().padStart(2,'0')}.${now.getMilliseconds().toString().padStart(3,'0')}`
}

export type LogFn = (message: string) => void;

const log: LogFn = (message) => {
  console.log(`${time()} ${message}`);
}

export default log;