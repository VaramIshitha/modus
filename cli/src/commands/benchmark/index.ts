/*
 * Copyright 2024 Hypermode Inc.
 * Licensed under the terms of the Apache License, Version 2.0
 * See the LICENSE file that accompanied this code for further details.
 *
 * SPDX-FileCopyrightText: 2024 Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

import { Args, Flags } from "@oclif/core";
import { spawn } from "node:child_process";
import path from "node:path";
import os from "node:os";
import chalk from "chalk";

import * as fs from "../../util/fs.js";
import * as vi from "../../util/versioninfo.js";
import * as installer from "../../util/installer.js";
import { getHeader } from "../../custom/header.js";
import { getAppInfo } from "../../util/appinfo.js";
import { isOnline, withSpinner } from "../../util/index.js";
import { readHypermodeSettings } from "../../util/hypermode.js";
import BuildCommand from "../build/index.js";
import { BaseCommand } from "../../baseCommand.js";
import { execFile } from "../../util/cp.js";
import * as inquirer from "@inquirer/prompts";
import { gql, GraphQLClient } from "graphql-request";
import { readFile } from "fs/promises";
import { parse } from "csv-parse/sync";

const MANIFEST_FILE = "modus.json";
const ENV_FILES = [".env", ".env.local", ".env.development", ".env.dev", ".env.development.local", ".env.dev.local"];

export default class BenchmarkCommand extends BaseCommand {
  static args = {
    path: Args.directory({
      description: "Path to app directory",
      default: ".",
      exists: true,
    }),
    input: Args.file({
      description: "Path to input CSV file",
      default: "./input.csv",
    }),
  };

  static flags = {
    runtime: Flags.string({
      char: "r",
      description: "Modus runtime version to use. If not provided, the latest runtime compatible with the app will be used.",
    }),
    prerelease: Flags.boolean({
      char: "p",
      aliases: ["pre"],
      description: "Use a prerelease version of the Modus runtime. Not needed if specifying a runtime version.",
    }),
    delay: Flags.integer({
      description: "Delay (in milliseconds) between file change detection and rebuild",
      default: 500,
    }),
  };

  static description = "Benchmark multiple Modus apps locally";

  static examples = ["modus benchmark"];

  async run(): Promise<void> {
    const { args, flags } = await this.parse(BenchmarkCommand);

    const appPath = path.resolve(args.path);
    if (!(await fs.exists(path.join(appPath, MANIFEST_FILE)))) {
      this.log(chalk.red(`A ${MANIFEST_FILE} file was not found at ${appPath}`));
      this.log(chalk.red("Please either execute the modus command from the app directory, or specify the path to the app you want to run."));
      this.exit(1);
    }

    const app = await getAppInfo(appPath);
    const { sdk, sdkVersion } = app;

    if (!flags["no-logo"]) {
      this.log(getHeader(this.config.version));
    }

    if (!(await vi.sdkVersionIsInstalled(sdk, sdkVersion))) {
      const sdkText = `Modus ${sdk} SDK ${sdkVersion}`;
      await withSpinner(chalk.dim("Downloading and installing " + sdkText), async (spinner) => {
        try {
          await installer.installSDK(sdk, sdkVersion);
        } catch (e) {
          spinner.fail(chalk.red(`Failed to download ${sdkText}`));
          throw e;
        }
        spinner.succeed(chalk.dim(`Installed ${sdkText}`));
      });
    }

    let runtimeVersion = flags.runtime;
    if (runtimeVersion) {
      const runtimeText = `Modus Runtime ${runtimeVersion}`;
      if (!(await vi.runtimeVersionIsInstalled(runtimeVersion))) {
        if (await isOnline()) {
          await withSpinner(chalk.dim("Downloading and installing " + runtimeText), async (spinner) => {
            try {
              await installer.installRuntime(runtimeVersion!);
            } catch (e) {
              spinner.fail(chalk.red("Failed to download " + runtimeText));
              throw e;
            }
            spinner.succeed(chalk.dim("Installed " + runtimeText));
          });
        } else {
          this.logError(`${runtimeText} is not installed, and you are offline. Please try again when you have an internet connection.`);
          this.exit(1);
        }
      }
    } else if (await isOnline()) {
      const version = await vi.findLatestCompatibleRuntimeVersion(sdk, sdkVersion, flags.prerelease);
      if (version && !(await vi.runtimeVersionIsInstalled(version))) {
        const runtimeText = `Modus Runtime ${version}`;
        await withSpinner(chalk.dim("Downloading and installing " + runtimeText), async (spinner) => {
          try {
            await installer.installRuntime(version!);
          } catch (e) {
            spinner.fail(chalk.red("Failed to download " + runtimeText));
            throw e;
          }
          spinner.succeed(chalk.dim("Installed " + runtimeText));
        });
      }
      if (!version) {
        this.logError("Could not find a compatible Modus runtime version. Please try again.");
        return;
      }
      runtimeVersion = version;
    } else {
      const version = await vi.findCompatibleInstalledRuntimeVersion(sdk, sdkVersion, flags.prerelease);
      if (!version) {
        this.logError("Could not find a compatible Modus runtime version. Please try again when you have an internet connection.");
        return;
      }
      runtimeVersion = version;
    }

    const ext = os.platform() === "win32" ? ".exe" : "";
    const runtimePath = path.join(vi.getRuntimePath(runtimeVersion), "modus_runtime" + ext);

    if (await fs.exists(path.join(appPath, "build"))) {
      await fs.rm(path.join(appPath, "build"), { recursive: true });
    }

    const branches = await getGitBranches(appPath);
    const selectedBranches = await inquirer.checkbox({
      message: "Select branches to benchmark",
      choices: branches.map((b) => {
        return {
          name: b,
          value: b,
        };
      }),
    });

    // Read Hypermode settings if they exist, so they can be forwarded to the runtime
    const hypSettings = await readHypermodeSettings();

    const env = {
      ...process.env,
      MODUS_ENV: "dev",
      HYP_EMAIL: hypSettings.email,
      HYP_JWT: hypSettings.jwt,
      HYP_ORG_ID: hypSettings.orgId,
    };

    const basePort = 8686;
    const branchesToPort = new Map<string, number>();
    selectedBranches.forEach((branch, i) => {
      branchesToPort.set(branch, basePort + i);
    });

    for (const branch of selectedBranches) {
      await switchToGitBranch(appPath, branch);
      const appBuildPath = path.join(appPath, "build", branch);
      await BuildCommand.run([appPath, "--output", appBuildPath, "--no-logo"]);

      // Copy env files to the build directory
      await copyEnvFiles(appPath, appBuildPath);

      // Spawn the runtime child process
      const child = spawn(runtimePath, ["-appPath", appBuildPath, "-port", branchesToPort.get(branch)!.toString(), "-refresh=1s"], {
        stdio: ["inherit", "inherit", "pipe"],
        env: env,
      });
      child.stderr.pipe(process.stderr);

      // Handle the runtime process exit
      child.on("close", (code) => {
        // note: can't use "this.exit" here because it would throw an unhandled exception
        // but "process.exit" works fine.
        if (code) {
          this.log(chalk.magentaBright(`Runtime terminated with code ${code}`) + "\n");
          process.exit(code);
        } else {
          this.log(chalk.magentaBright("Runtime terminated successfully.") + "\n");
          process.exit();
        }
      });

      // Forward SIGINT and SIGTERM to the child process for graceful shutdown from user ctrl+c or kill.
      process.on("SIGINT", () => {
        if (child && !child.killed) {
          child.kill("SIGINT");
        }
      });
      process.on("SIGTERM", () => {
        if (child && !child.killed) {
          child.kill("SIGTERM");
        }
      });
    }

    await runRequests(args.input, branchesToPort);
  }

  private logError(message: string) {
    this.log(chalk.red(" ERROR ") + chalk.dim(": " + message));
  }
}

async function runRequests(inputFile: string, branchesToPort: Map<string, number>) {
  const inputs = await readCSVFile(inputFile);

  const rows = [];

  const endpoint = "https://eval-kevinm.hypermode.app/graphql";
  const evalClient = new GraphQLClient(endpoint, {
    headers: {
      authorization: "Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjMxMTMxMjMsImlhdCI6MTczMTU3NzEyMywiaXNzIjoiaHlwZXJtb2RlLmNvbSIsInN1YiI6ImFway0wMTkzMmEwYS0xM2NjLTdkM2MtOTVhMy1lNTRhNzRhMWM2NTEifQ.5s7xxhA7L2VHmV0TIXTCWAFRAuaUtbh6gg_0PYTwymUshrrv5KhuNtrcoe6NQcNjshXdeAlwi6aJeWBPET837Q",
    },
  });

  const branchToClient = new Map<string, GraphQLClient>();
  for (const [branch, port] of branchesToPort) {
    const endpoint = `http://localhost:${port}/graphql`;
    branchToClient.set(branch, new GraphQLClient(endpoint));
  }

  for (const input of inputs) {
    const query = gql`
      query ($name: String!) {
        sayHello(name: $name)
      }
    `;

    const variables = {
      name: input,
    };

    try {
      const rowContent: any = {
        input,
      };

      for (const [branch, client] of branchToClient) {
        const response: any = await client.request(query, variables);
        rowContent[branch] = response.sayHello;

        const evalQuery = gql`
          query ScoreResponse($input: String!, $output: String!) {
            scoreResponse(input: $input, output: $output) {
              score
              reasoning
            }
          }
        `;

        const evalVariables = {
          input: rowContent[branch],
          output: response.sayHello,
        };

        const evalResponse: any = await evalClient.request(evalQuery, evalVariables);
        rowContent[branch + "score"] = evalResponse.scoreResponse.score;
      }

      rows.push(rowContent);
    } catch (error) {
      console.error(`Error`, error);
    }
  }

  writeJSONToCSV(rows, "./output.csv");
}

async function copyEnvFiles(appPath: string, buildPath: string): Promise<void> {
  for (const file of ENV_FILES) {
    const src = path.join(appPath, file);
    const dest = path.join(buildPath, file);
    if (await fs.exists(src)) {
      await fs.copyFile(src, dest);
    } else if (await fs.exists(dest)) {
      await fs.unlink(dest);
    }
  }
}

async function getGitBranches(cwd: string) {
  const execOpts = { env: process.env, cwd, shell: true };
  const { stdout: branchesStdOut } = await execFile("git", ["branch", `--format="%(refname:short)"`], execOpts);

  return branchesStdOut
    .trim()
    .split("\n")
    .map((b) => b.trim());
}

async function switchToGitBranch(cwd: string, branch: string) {
  const execOpts = { env: process.env, cwd, shell: true };
  await execFile("git", ["checkout", `${branch}`], execOpts);
}

async function readCSVFile(filePath: string): Promise<string[]> {
  try {
    const fileContent = await readFile(filePath, "utf8");

    const records = parse(fileContent, {
      columns: true, // Parse with headers as keys
      skip_empty_lines: true,
    });

    return records.map((r: any) => r.input);
  } catch (error) {
    console.error("Error reading CSV file:", error);
    return [];
  }
}

import { writeFile } from "fs/promises";
import { stringify } from "csv-stringify/sync";

async function writeJSONToCSV(jsonData: any[], outputPath: string) {
  try {
    // Convert the JSON data to CSV format
    const csvData = stringify(jsonData, {
      header: true, // Include header row
    });

    // Write the CSV string to a file
    await writeFile(outputPath, csvData);

    console.log(`CSV file successfully written to ${outputPath}`);
  } catch (error) {
    console.error("Error writing CSV file:", error);
  }
}
