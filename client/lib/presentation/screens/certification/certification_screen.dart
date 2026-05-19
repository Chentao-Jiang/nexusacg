import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexusacg/core/network/api_client.dart';

enum CertType { merchant, serviceProvider }

enum ProviderType {
  makeupArtist('化妆造型'),
  wigStylist('假发造型'),
  photographer('摄影跟拍'),
  postEditor('后期修图'),
  propsMaker('道具制作');

  final String label;
  const ProviderType(this.label);
}

class CertificationScreen extends StatefulWidget {
  const CertificationScreen({super.key});

  @override
  State<CertificationScreen> createState() => _CertificationScreenState();
}

class _CertificationScreenState extends State<CertificationScreen> {
  final _api = ApiClient();
  final _formKey = GlobalKey<FormState>();
  final _storeNameCtrl = TextEditingController();
  final _descCtrl = TextEditingController();

  CertType _certType = CertType.merchant;
  ProviderType? _selectedProviderType;
  File? _licenseFile;
  final List<File> _portfolioFiles = [];
  bool _submitting = false;

  @override
  void dispose() {
    _storeNameCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickLicense() async {
    final picked = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 85,
    );
    if (picked != null) {
      setState(() => _licenseFile = File(picked.path));
    }
  }

  Future<void> _pickPortfolio() async {
    final picked = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 85,
    );
    if (picked != null) {
      final file = File(picked.path);
      if (!mounted) return;
      if (_portfolioFiles.length >= 10) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('最多上传10张作品图片')),
        );
        return;
      }
      setState(() => _portfolioFiles.add(File(picked.path)));
    }
  }

  void _removePortfolio(int index) {
    setState(() => _portfolioFiles.removeAt(index));
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_certType == CertType.merchant && _licenseFile == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请上传营业执照')),
      );
      return;
    }
    if (_certType == CertType.serviceProvider && _selectedProviderType == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请选择服务类型')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      String? licenseUrl;
      if (_licenseFile != null) {
        licenseUrl = await _api.uploadImage(_licenseFile!);
      }

      final uploadedPortfolio = <String>[];
      for (final file in _portfolioFiles) {
        final url = await _api.uploadImage(file);
        if (url != null) uploadedPortfolio.add(url);
      }

      final Map<String, dynamic> body;
      if (_certType == CertType.merchant) {
        body = {
          'product_category': 'cosplay',
          'store_name': _storeNameCtrl.text.trim(),
          if (licenseUrl != null) 'business_license_url': licenseUrl,
        };
      } else {
        body = {
          'provider_type': _providerTypeToApiValue(_selectedProviderType!),
          'description': _descCtrl.text.trim(),
          'portfolio_images': uploadedPortfolio,
          if (licenseUrl != null) 'business_license_url': licenseUrl,
        };
      }

      final endpoint = _certType == CertType.merchant
          ? '/certifications/merchant'
          : '/certifications/service-provider';

      final response = await _api.post(endpoint, data: body);
      final data = response.data;
      if (mounted) {
        if (data is Map && data['code'] == 0) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('认证申请已提交，等待审核')),
          );
          Navigator.pop(context);
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(data is Map && data['message'] != null
                ? data['message'] as String
                : '提交失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('提交失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  String _providerTypeToApiValue(ProviderType type) {
    switch (type) {
      case ProviderType.makeupArtist: return 'makeup_artist';
      case ProviderType.wigStylist: return 'wig_stylist';
      case ProviderType.photographer: return 'photographer';
      case ProviderType.postEditor: return 'post_editor';
      case ProviderType.propsMaker: return 'props_maker';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('商家入驻')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SegmentedButton<CertType>(
                segments: const [
                  ButtonSegment(value: CertType.merchant,
                    label: Text('商家入驻'), icon: Icon(Icons.store)),
                  ButtonSegment(value: CertType.serviceProvider,
                    label: Text('服务者入驻'), icon: Icon(Icons.person)),
                ],
                selected: {_certType},
                onSelectionChanged: (s) => setState(() => _certType = s.first),
              ),
              const SizedBox(height: 24),
              if (_certType == CertType.merchant) ...[
                TextFormField(
                  controller: _storeNameCtrl,
                  decoration: const InputDecoration(
                    labelText: '店铺名称',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => v == null || v.isEmpty ? '请输入店铺名称' : null,
                ),
                const SizedBox(height: 16),
                const Text('营业执照', style: TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                GestureDetector(
                  onTap: _pickLicense,
                  child: Container(
                    height: 150,
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.grey.shade300),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: _licenseFile != null
                        ? ClipRRect(
                            borderRadius: BorderRadius.circular(8),
                            child: Image.file(_licenseFile!, fit: BoxFit.cover),
                          )
                        : const Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.add_a_photo, size: 40, color: Colors.grey),
                              SizedBox(height: 8),
                              Text('点击上传营业执照', style: TextStyle(color: Colors.grey)),
                            ],
                          ),
                  ),
                ),
              ] else ...[
                const Text('服务类型', style: TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                DropdownButtonFormField<ProviderType>(
                  value: _selectedProviderType,
                  decoration: const InputDecoration(border: OutlineInputBorder()),
                  items: ProviderType.values.map((t) =>
                    DropdownMenuItem(value: t, child: Text(t.label)),
                  ).toList(),
                  onChanged: (v) => setState(() => _selectedProviderType = v),
                  validator: (v) => v == null ? '请选择服务类型' : null,
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _descCtrl,
                  decoration: const InputDecoration(
                    labelText: '服务描述',
                    border: OutlineInputBorder(),
                    alignLabelWithHint: true,
                  ),
                  maxLines: 4,
                  maxLength: 500,
                  validator: (v) => v == null || v.isEmpty ? '请输入服务描述' : null,
                ),
                const SizedBox(height: 16),
                const Text('作品图片（最多10张）', style: TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    ...List.generate(_portfolioFiles.length, (i) =>
                      Stack(
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(8),
                            child: Image.file(_portfolioFiles[i],
                                width: 100, height: 100, fit: BoxFit.cover),
                          ),
                          Positioned(
                            top: 2, right: 2,
                            child: GestureDetector(
                              onTap: () => _removePortfolio(i),
                              child: Container(
                                padding: const EdgeInsets.all(2),
                                decoration: const BoxDecoration(
                                  color: Colors.black54, shape: BoxShape.circle),
                                child: const Icon(Icons.close, size: 14, color: Colors.white),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (_portfolioFiles.length < 10)
                      GestureDetector(
                        onTap: _pickPortfolio,
                        child: Container(
                          width: 100, height: 100,
                          decoration: BoxDecoration(
                            border: Border.all(color: Colors.grey.shade300),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: const Icon(Icons.add_a_photo, color: Colors.grey),
                        ),
                      ),
                  ],
                ),
              ],
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _submitting ? null : _submit,
                  child: _submitting
                      ? const SizedBox(height: 20, width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : const Text('提交申请'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
