const a = new Error("Hello");

if (!a instanceof Error) {
  console.log("surprising");
}
