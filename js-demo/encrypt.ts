import * as forge from "node-forge";

/**
 * 简单的 SHA256 加密函数
 */
function sha256Simple(data: string): string {
  const md = forge.md.sha256.create();
  md.update(data);
  return md.digest().toHex();
}

// 测试
const testString = "Quebec@123456";
const hash = sha256Simple(testString);
console.log(`"${testString}" 的 SHA256:`);
console.log(hash);
console.log("09b9b136dd7047f774ce919ed7cd7a1045dce4c773f33fdf83922b43589bac2e");
