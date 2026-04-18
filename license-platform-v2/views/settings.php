<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">系统设置</h1>

    <?php if (!empty($success)): ?>
    <div class="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded"><?= htmlspecialchars($success) ?></div>
    <?php endif; ?>
    <?php if (!empty($error)): ?>
    <div class="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded"><?= htmlspecialchars($error) ?></div>
    <?php endif; ?>

    <!-- Tab 导航 -->
    <div class="border-b border-gray-200 mb-6">
        <nav class="-mb-px flex space-x-8">
            <button onclick="showTab('pricing')" id="tab-pricing" class="tab-btn border-indigo-500 text-indigo-600 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm">价格方案</button>
            <button onclick="showTab('payment')" id="tab-payment" class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm">支付配置</button>
            <button onclick="showTab('email')" id="tab-email" class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm">邮件通知</button>
            <button onclick="showTab('security')" id="tab-security" class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm">登录警报</button>
        </nav>
    </div>

    <script>
    function showTab(name) {
        document.querySelectorAll('.tab-content').forEach(el => el.classList.add('hidden'));
        document.querySelectorAll('.tab-btn').forEach(el => { el.className = el.className.replace(/border-indigo-500 text-indigo-600/g, 'border-transparent text-gray-500'); });
        document.getElementById('panel-' + name).classList.remove('hidden');
        document.getElementById('tab-' + name).className = document.getElementById('tab-' + name).className.replace('border-transparent text-gray-500', 'border-indigo-500 text-indigo-600');
    }
    </script>

    <!-- 价格方案 -->
    <div id="panel-pricing" class="tab-content">
        <form method="POST" action="/admin/settings" class="space-y-4">
            <input type="hidden" name="section" value="pricing">
            <div id="tier-list" class="space-y-4">
                <?php $i = 0; foreach ($tiers as $key => $tier): ?>
                <div class="bg-white rounded-lg shadow p-4 tier-row">
                    <div class="grid grid-cols-1 md:grid-cols-5 gap-4 items-end">
                        <div>
                            <label class="block text-xs text-gray-500">套餐标识</label>
                            <input type="text" name="tier_key[]" value="<?= htmlspecialchars($key) ?>"
                                class="w-full border rounded p-2 text-sm" placeholder="如 pro">
                        </div>
                        <div>
                            <label class="block text-xs text-gray-500">显示名称</label>
                            <input type="text" name="tier_name[]" value="<?= htmlspecialchars($tier['name'] ?? '') ?>"
                                class="w-full border rounded p-2 text-sm" placeholder="如 Pro">
                        </div>
                        <div>
                            <label class="block text-xs text-gray-500">月价格($)</label>
                            <input type="number" name="tier_price[]" value="<?= $tier['price'] ?? 0 ?>" min="0"
                                class="w-full border rounded p-2 text-sm">
                        </div>
                        <div>
                            <label class="block text-xs text-gray-500">最大设备数</label>
                            <input type="number" name="tier_max_devices[]" value="<?= $tier['max_devices'] ?? 1 ?>" min="1"
                                class="w-full border rounded p-2 text-sm">
                        </div>
                        <div>
                            <label class="block text-xs text-gray-500">功能列表(逗号分隔)</label>
                            <input type="text" name="tier_features[]" value="<?= htmlspecialchars(implode(',', $tier['features'] ?? [])) ?>"
                                class="w-full border rounded p-2 text-sm" placeholder="graph,backup">
                        </div>
                    </div>
                    <button type="button" onclick="this.closest('.tier-row').remove()" class="mt-2 text-red-500 text-xs hover:text-red-700">删除此套餐</button>
                </div>
                <?php $i++; endforeach; ?>
            </div>
            <button type="button" onclick="addTier()" class="text-indigo-600 text-sm hover:text-indigo-800">+ 添加套餐</button>
            <div class="pt-4">
                <button type="submit" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700">保存价格方案</button>
            </div>
        </form>
        <script>
        function addTier() {
            const html = `<div class="bg-white rounded-lg shadow p-4 tier-row">
                <div class="grid grid-cols-1 md:grid-cols-5 gap-4 items-end">
                    <div><label class="block text-xs text-gray-500">套餐标识</label><input type="text" name="tier_key[]" class="w-full border rounded p-2 text-sm" placeholder="custom"></div>
                    <div><label class="block text-xs text-gray-500">显示名称</label><input type="text" name="tier_name[]" class="w-full border rounded p-2 text-sm" placeholder="Custom"></div>
                    <div><label class="block text-xs text-gray-500">月价格($)</label><input type="number" name="tier_price[]" value="0" min="0" class="w-full border rounded p-2 text-sm"></div>
                    <div><label class="block text-xs text-gray-500">最大设备数</label><input type="number" name="tier_max_devices[]" value="1" min="1" class="w-full border rounded p-2 text-sm"></div>
                    <div><label class="block text-xs text-gray-500">功能列表(逗号分隔)</label><input type="text" name="tier_features[]" class="w-full border rounded p-2 text-sm" placeholder="graph,backup"></div>
                </div>
                <button type="button" onclick="this.closest('.tier-row').remove()" class="mt-2 text-red-500 text-xs hover:text-red-700">删除此套餐</button>
            </div>`;
            document.getElementById('tier-list').insertAdjacentHTML('beforeend', html);
        }
        </script>
    </div>

    <!-- 支付配置 -->
    <div id="panel-payment" class="tab-content hidden">
        <form method="POST" action="/admin/settings" class="space-y-6 bg-white rounded-lg shadow p-6">
            <input type="hidden" name="section" value="payment">
            
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">支付方式</label>
                <select name="payment_method" class="rounded-md border-gray-300 shadow-sm border p-2 w-full">
                    <option value="manual" <?= $paymentMethod === 'manual' ? 'selected' : '' ?>>手动处理（线下转账）</option>
                    <option value="alipay" <?= $paymentMethod === 'alipay' ? 'selected' : '' ?>>支付宝</option>
                    <option value="wechat" <?= $paymentMethod === 'wechat' ? 'selected' : '' ?>>微信支付</option>
                </select>
                <p class="mt-1 text-xs text-gray-500">选择"手动处理"时，用户购买后由管理员手动创建授权码</p>
            </div>

            <div class="border-t pt-4">
                <h3 class="text-sm font-medium text-gray-900 mb-3">支付宝配置</h3>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs text-gray-500">App ID</label>
                        <input type="text" name="alipay_app_id" value="<?= htmlspecialchars($allSettings['alipay_app_id'] ?? '') ?>"
                            class="w-full border rounded p-2 text-sm">
                    </div>
                    <div>
                        <label class="block text-xs text-gray-500">应用私钥</label>
                        <textarea name="alipay_private_key" rows="2" class="w-full border rounded p-2 text-sm"><?= htmlspecialchars($allSettings['alipay_private_key'] ?? '') ?></textarea>
                    </div>
                    <div>
                        <label class="block text-xs text-gray-500">支付宝公钥</label>
                        <textarea name="alipay_public_key" rows="2" class="w-full border rounded p-2 text-sm"><?= htmlspecialchars($allSettings['alipay_public_key'] ?? '') ?></textarea>
                    </div>
                </div>
            </div>

            <div class="border-t pt-4">
                <h3 class="text-sm font-medium text-gray-900 mb-3">微信支付配置</h3>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs text-gray-500">商户号</label>
                        <input type="text" name="wechat_mch_id" value="<?= htmlspecialchars($allSettings['wechat_mch_id'] ?? '') ?>"
                            class="w-full border rounded p-2 text-sm">
                    </div>
                    <div>
                        <label class="block text-xs text-gray-500">API密钥</label>
                        <input type="text" name="wechat_api_key" value="<?= htmlspecialchars($allSettings['wechat_api_key'] ?? '') ?>"
                            class="w-full border rounded p-2 text-sm">
                    </div>
                    <div>
                        <label class="block text-xs text-gray-500">App ID</label>
                        <input type="text" name="wechat_app_id" value="<?= htmlspecialchars($allSettings['wechat_app_id'] ?? '') ?>"
                            class="w-full border rounded p-2 text-sm">
                    </div>
                </div>
            </div>

            <button type="submit" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700">保存支付配置</button>
        </form>
    </div>

    <!-- 邮件通知 -->
    <div id="panel-email" class="tab-content hidden">
        <form method="POST" action="/admin/settings" class="space-y-4 bg-white rounded-lg shadow p-6">
            <input type="hidden" name="section" value="email">
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <label class="block text-xs text-gray-500">SMTP 服务器</label>
                    <input type="text" name="smtp_host" value="<?= htmlspecialchars($allSettings['smtp_host'] ?? '') ?>"
                        class="w-full border rounded p-2 text-sm" placeholder="smtp.qq.com">
                </div>
                <div>
                    <label class="block text-xs text-gray-500">SMTP 端口</label>
                    <input type="text" name="smtp_port" value="<?= htmlspecialchars($allSettings['smtp_port'] ?? '465') ?>"
                        class="w-full border rounded p-2 text-sm">
                </div>
                <div>
                    <label class="block text-xs text-gray-500">SMTP 用户名</label>
                    <input type="text" name="smtp_user" value="<?= htmlspecialchars($allSettings['smtp_user'] ?? '') ?>"
                        class="w-full border rounded p-2 text-sm">
                </div>
                <div>
                    <label class="block text-xs text-gray-500">SMTP 密码/授权码</label>
                    <input type="password" name="smtp_pass" value="<?= htmlspecialchars($allSettings['smtp_pass'] ?? '') ?>"
                        class="w-full border rounded p-2 text-sm">
                </div>
                <div>
                    <label class="block text-xs text-gray-500">发件人地址</label>
                    <input type="email" name="smtp_from" value="<?= htmlspecialchars($allSettings['smtp_from'] ?? '') ?>"
                        class="w-full border rounded p-2 text-sm">
                </div>
                <div>
                    <label class="block text-xs text-gray-500">到期提前提醒天数</label>
                    <input type="number" name="expiry_notify_days" value="<?= htmlspecialchars($allSettings['expiry_notify_days'] ?? '7') ?>"
                        min="1" max="30" class="w-full border rounded p-2 text-sm">
                </div>
            </div>
            <p class="text-xs text-gray-500">配置 SMTP 后，系统会在授权码到期前自动发送邮件提醒客户。客户邮箱在创建授权码时填写。</p>

            <button type="submit" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700">保存邮件配置</button>
        </form>
    </div>

    <!-- 登录警报 -->
    <div id="panel-security" class="tab-content hidden">
        <div class="bg-white rounded-lg shadow p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">登录日志（最近50条）</h3>
            <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
                            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">邮箱</th>
                            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP</th>
                            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">浏览器</th>
                            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">结果</th>
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200">
                        <?php foreach ($loginLogs as $log): ?>
                        <tr class="<?= $log['success'] ? '' : 'bg-red-50' ?>">
                            <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-500">
                                <?= date('m-d H:i', strtotime($log['created_at'])) ?>
                            </td>
                            <td class="px-4 py-2 whitespace-nowrap text-sm"><?= htmlspecialchars($log['email']) ?></td>
                            <td class="px-4 py-2 whitespace-nowrap text-sm font-mono"><?= htmlspecialchars($log['ip_address']) ?></td>
                            <td class="px-4 py-2 text-xs text-gray-400 max-w-[200px] truncate"><?= htmlspecialchars(substr($log['user_agent'] ?? '', 0, 60)) ?></td>
                            <td class="px-4 py-2 whitespace-nowrap">
                                <?php if ($log['success']): ?>
                                <span class="px-2 py-0.5 text-xs font-medium rounded-full bg-green-100 text-green-800">成功</span>
                                <?php else: ?>
                                <span class="px-2 py-0.5 text-xs font-medium rounded-full bg-red-100 text-red-800">失败</span>
                                <?php endif; ?>
                            </td>
                        </tr>
                        <?php endforeach; ?>
                        <?php if (empty($loginLogs)): ?>
                        <tr><td colspan="5" class="px-4 py-4 text-center text-gray-500">暂无记录</td></tr>
                        <?php endif; ?>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>
