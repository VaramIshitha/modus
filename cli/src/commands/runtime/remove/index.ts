/*
 * Copyright 2024 Hypermode Inc.
 * Licensed under the terms of the Apache License, Version 2.0
 * See the LICENSE file that accompanied this code for further details.
 *
 * SPDX-FileCopyrightText: 2024 Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

import { Args, Command, Flags } from "@oclif/core";
import chalk from "chalk";

import * as fs from "../../../util/fs.js";
import * as vi from "../../../util/versioninfo.js";
import { ask, clearLine, withSpinner } from "../../../util/index.js";

export default class RuntimeRemoveCommand extends Command {
  static args = {
    version: Args.string({
      description: "Runtime version to remove, or 'all' to remove all runtimes.",
      required: true,
    }),
  };

  static flags = {
    force: Flags.boolean({
      char: "f",
      default: false,
      description: "Remove without prompting",
    }),
  };

  static description = "Remove a Modus runtime";
  static examples = ["modus runtime remove v0.0.0", "modus runtime remove all"];

  async run(): Promise<void> {
    const { args, flags } = await this.parse(RuntimeRemoveCommand);
    if (!args.version) {
      this.logError(`No runtime version specified! Run ${chalk.whiteBright("modus runtime remove <version>")}, or ${chalk.whiteBright("modus runtime remove all")}`);
      return;
    }

    if (args.version.toLowerCase() === "all") {
      const versions = await vi.getInstalledRuntimeVersions();
      if (versions.length === 0) {
        this.log(chalk.yellow("No Modus runtimes are installed."));
        this.exit(1);
      } else if (!flags.force && !(await this.confirmAction("Really, remove all Modus runtimes? [y/n]"))) {
        this.log(chalk.dim("Aborted."));
        this.exit(1);
      }

      for (const version of versions) {
        await this.removeRuntime(version);
      }
    } else if (!args.version.startsWith("v")) {
      this.logError("Version must start with 'v'.");
      this.exit(1);
    } else {
      const runtimeText = `Modus Runtime ${args.version}`;
      const isInstalled = await vi.runtimeVersionIsInstalled(args.version);
      if (!isInstalled) {
        this.log(chalk.yellow(runtimeText + "is not installed."));
        this.exit(1);
      } else if (!flags.force && !(await this.confirmAction(`Really, remove ${runtimeText} ? [y/n]`))) {
        this.log(chalk.dim("Aborted."));
        this.exit(1);
      }

      await this.removeRuntime(args.version);
    }
  }

  private async removeRuntime(version: string): Promise<void> {
    const runtimeText = `Modus Runtime ${version}`;
    await withSpinner(chalk.dim("Removing " + runtimeText), async (spinner) => {
      const dir = vi.getRuntimePath(version);
      try {
        await fs.rm(dir, { recursive: true, force: true });
        spinner.succeed(chalk.dim("Removed " + runtimeText));
      } catch (e) {
        spinner.fail(chalk.red("Failed to remove " + runtimeText));
        throw e;
      }
    });
  }

  private logError(message: string) {
    this.log(chalk.red(" ERROR ") + chalk.dim(": " + message));
  }

  private async confirmAction(message: string): Promise<boolean> {
    this.log(message);
    const cont = ((await ask(chalk.dim(" -> "))) || "n").toLowerCase().trim();
    clearLine(2);
    return cont === "yes" || cont === "y";
  }
}