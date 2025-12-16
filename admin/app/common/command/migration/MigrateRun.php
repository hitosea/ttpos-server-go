<?php
// +----------------------------------------------------------------------
// | TopThink [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2016 http://www.topthink.com All rights reserved.
// +----------------------------------------------------------------------
// | Author: zhangyajun <448901948@qq.com>
// +----------------------------------------------------------------------

namespace app\common\command\migration;

use think\facade\Db;
use think\console\Input;
use think\facade\Config;
use think\console\Output;
use think\migration\command\Migrate;
use Phinx\Migration\MigrationInterface;
use think\console\input\Option as InputOption;

class MigrateRun extends Migrate
{

    /**
     * {@inheritdoc}
     */
    protected function configure()
    {
        $this->setName('migrate:run')
            ->setDescription('Migrate the database')
            ->addOption('--target', '-t', InputOption::VALUE_REQUIRED, 'The version number to migrate to')
            ->addOption('--date', '-d', InputOption::VALUE_REQUIRED, 'The date to migrate to')
            ->addOption('--db', '-db', InputOption::VALUE_OPTIONAL, '指定数据库')
            ->addOption('--dry-run', '-dr', InputOption::VALUE_NONE, 'Dry run')
            ->setHelp(
                <<<EOT
The <info>migrate:run</info> command runs all available migrations, optionally up to a specific version

<info>php think migrate:run</info>
<info>php think migrate:run -t 20110103081132</info>
<info>php think migrate:run -d 20110103</info>
<info>php think migrate:run -v</info>

EOT
            );
    }

    /**
     * Migrate the database.
     *
     * @param Input  $input
     * @param Output $output
     */
    protected function execute(Input $input, Output $output)
    {
        $version = $input->getOption('target');
        $dbname = $input->getOption('db');
        // run the migrations
        $start = microtime(true);
        //
        $default = config('database.default');
        $config = config('database');
        $mysql = $config['connections'][$default];
        $tables = Db::getTables();
        $dbs = Db::query('SHOW DATABASES LIKE "shop\_%"');
        if (in_array($mysql['prefix'] . 'company', $tables)) {

            if ($dbname) {
                $mysql['database'] = $dbname;
                $mysql['username'] = env('DB_USERNAME');
                $mysql['password'] = env('DB_PASSWORD');
                $config['connections'][$default] = $mysql;
                Config::set($config, 'database');
                //
                $this->adapter = null;
                $this->migrations = null;
                $this->migrate($version, 2);
            } else {
                //
                $this->migrate($version, 1);
                //
                foreach (Db::name('company')->where('delete_time', 0)->column('uuid') as $appid) {
                    $mysql['database'] = 'shop' . $appid;
                    $mysql['username'] = env('DB_USERNAME');
                    $mysql['password'] = env('DB_PASSWORD');
                    $config['connections'][$default] = $mysql;
                    Config::set($config, 'database');
                    //
                    $this->adapter = null;
                    $this->migrations = null;
                    $this->migrate($version, 2);
                }
                // 连锁店
                $shopMasterList = [];
                foreach ($dbs as $db) {
                    if (strpos($db['Database (shop\_%)'], 'shop_') === 0) {
                        $shopMasterList[] = $db['Database (shop\_%)'];
                    }
                }
                if ($shopMasterList) {
                    foreach ($shopMasterList as $shopMaster) {
                        $mysql['database'] = $shopMaster;
                        $mysql['username'] = env('DB_USERNAME');
                        $mysql['password'] = env('DB_PASSWORD');
                        $config['connections'][$default] = $mysql;
                        Config::set($config, 'database');
                        //
                        $this->adapter = null;
                        $this->migrations = null;
                        $this->migrate($version, 3);
                    }
                }
            }
        } else {
            $this->migrate($version, 1);
        }
        //
        $end = microtime(true);
        $output->writeln('');
        $output->writeln('<comment>All Done. Took ' . sprintf('%.4fs', $end - $start) . '</comment>');
    }

    public function migrateToDateTime(\DateTime $dateTime)
    {
        $versions   = array_keys($this->getMigrations());
        $dateString = $dateTime->format('YmdHis');

        $outstandingMigrations = array_filter($versions, function ($version) use ($dateString) {
            return $version <= $dateString;
        });

        if (count($outstandingMigrations) > 0) {
            $migration = max($outstandingMigrations);
            $this->output->writeln('Migrating to version ' . $migration);
            $this->migrate($migration);
        }
    }

    protected function migrate($version = null, $mode = 1)
    {
        $migrations = $this->getMigrations();
        $versions   = $this->getVersions();
        $current    = $this->getCurrentVersion();

        if (empty($versions) && empty($migrations)) {
            return;
        }

        if (null === $version) {
            $version = max(array_merge($versions, array_keys($migrations)));
        } else {
            if (0 != $version && !isset($migrations[$version])) {
                $this->output->writeln(sprintf('<comment>warning</comment> %s is not a valid version', $version));
                return;
            }
        }

        // are we migrating up or down?
        $direction = $version > $current ? MigrationInterface::UP : MigrationInterface::DOWN;


        if ($direction === MigrationInterface::DOWN) {
            // run downs first
            krsort($migrations);
            foreach ($migrations as $migration) {
                if ($migration->getVersion() <= $version) {
                    break;
                }

                if (in_array($migration->getVersion(), $versions)) {
                    $target = "shop";
                    try {
                        $target = $migration::TARGET;
                    } catch (\Throwable $th) {
                    }
                    // 优化后的迁移执行逻辑
                    if ($target == 'shop_master') {
                        // shop_master 只在 mode 为 2 或 3 时执行
                        if (in_array($mode, [2, 3])) {
                            $this->executeMigration($migration, MigrationInterface::DOWN);
                        }
                    } else if ($mode != 3) {
                        // mode 不为 3 时，判断目标类型
                        $isAll = $target == 'all';
                        $isMain = ($mode == 1 && $target == 'main');
                        $isShop = ($mode != 1 && $target == 'shop');
                        if ($isAll || $isMain || $isShop) {
                            $this->executeMigration($migration, MigrationInterface::DOWN);
                        }
                    }
                }
            }
        }

        ksort($migrations);
        foreach ($migrations as $migration) {
            if ($migration->getVersion() > $version) {
                break;
            }
            if (!in_array($migration->getVersion(), $versions)) {
                $target = "shop";
                try {
                    $target = $migration::TARGET;
                } catch (\Throwable $th) {
                }
                // 优化后的迁移执行逻辑
                if ($target == 'shop_master') {
                    // shop_master 只在 mode 为 2 或 3 时执行
                    if (in_array($mode, [2, 3])) {
                        $this->executeMigration($migration, MigrationInterface::UP);
                    }
                } else if ($mode != 3) {
                    // mode 不为 3 时，判断目标类型
                    $isAll = $target == 'all';
                    $isMain = ($mode == 1 && $target == 'main');
                    $isShop = ($mode != 1 && $target == 'shop');
                    if ($isAll || $isMain || $isShop) {
                        $this->executeMigration($migration, MigrationInterface::UP);
                    }
                }
            }
        }
    }

    protected function getCurrentVersion()
    {
        $versions = $this->getVersions();
        $version  = 0;

        if (!empty($versions)) {
            $version = end($versions);
        }

        return $version;
    }
}
