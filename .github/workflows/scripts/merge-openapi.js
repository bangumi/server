module.exports = async ({ github, exec }) => {
  const newBranch = "update-new-api";
  const owner = "bangumi";
  const branch = "master";
  const repo = "api";

  process.chdir("api");

  const options = { silent: true };
  try {
    await exec.exec("git ditt --exit-code", undefined, options);
    console.log("nothing to do");
    return;
  } catch (e) {
    console.log("should create pr for this");
  }

  await exec.exec(
    'git config --global user.email "github-action@users.noreply.github.com"'
  );
  await exec.exec('git config --global user.name "GitHub Action"');
  await exec.exec("git add .");

  await exec.exec(`git checkout -b ${newBranch}`);
  await exec.exec(`git commit -m "Update Openapi"`);

  try {
    await exec.exec(`git diff ${newBranch} origin/${branch} --exit-code`);
    console.log("nothing changed");
    return;
  } catch {
    await exec.exec(`git push origin ${newBranch} -f`);
  }

  const result = await github.rest.pulls.list({
    repo,
    owner,
    base: branch,
    head: `${owner}:${newBranch}`,
  });

  if (result.data.length === 0) {
    console.log("creating pr");
    await github.rest.pulls.create({
      repo,
      owner,
      base: branch,
      head: newBranch,
      title: "Update Openapi",
    });
  }
};
