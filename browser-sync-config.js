const path = require("path");
const fs = require("fs");

if (!process.env.PUBLIC_DIR) {
  throw new Error("PUBLIC_DIR environment variable is required");
}
const baseDir = process.env.PUBLIC_DIR;

module.exports = {
  server: {
    baseDir,
    middleware: [
      function (req, res, next) {
        const url = req.url.split("?")[0];
        if (path.extname(url)) {
          next();
          return;
        }
        const candidates = [
          path.join(baseDir, url),
          path.join(baseDir, url, "index")
        ];
        for (const c of candidates) {
          if (fs.existsSync(c) && fs.statSync(c).isFile()) {
            res.setHeader("Content-Type", "text/html; charset=utf-8");
            fs.createReadStream(c).pipe(res);
            return;
          }
        }
        next();
      }
    ]
  },
  files: [`${baseDir}/**/*`]
};
