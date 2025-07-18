
using Core;
using Microsoft.Maui.Controls;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;
using System.Text;
using System.Runtime.InteropServices;
using oadin_checker_0;

namespace FreeChat
{
    public partial class MainPage : ContentPage
    {
        string RespondContent;
        ChatAI ChatBot = new ChatAI();
        Embedding EmbeddingProcessor = new Embedding();
        TextToImageClient TextToImageClient = new TextToImageClient();
        string SettingsFilePath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "settings.txt");
        string userFolder = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);

        public MainPage()
        {
            string logPath = Path.Combine(userFolder, "oadin_log.txt");
            InitializeComponent();
            LoadSettings();

            InitializeOadin();
        }

        private async void InitializeOadin()
        {
            await OadinChecker.OadinInit(this);
        }

        // Navigation
        private void OnSelectionChanged(object sender, SelectionChangedEventArgs e)
        {
            var selectedOption = e.CurrentSelection[0].ToString();

            ChatView.IsVisible = selectedOption == "Chat";
            FileFrame.IsVisible = selectedOption == "File";
            TextToImageFrame.IsVisible = selectedOption == "TextToImage";
            SettingsFrame.IsVisible = selectedOption == "Settings";
        }

        // Chat interface
        private async void OnSend(Object sender, EventArgs e)
        {
            string userInput = PromptText.Text;

            if (!string.IsNullOrWhiteSpace(userInput))
            {
                Dialog.Children.Add(CreateChatBubble(userInput, isUser: true));

                PromptText.Text = string.Empty;

                var loadingBubble = CreateChatBubble(string.Empty, isUser: false);
                Dialog.Children.Add(loadingBubble);

                ChatBot.ChatHistory.Add(new { role = "user", content = userInput });

                var client = new HttpClient();
                var requestPayload = new
                {
                    model = ChatBot.ModelName,
                    messages = ChatBot.ChatHistory,
                    stream = true 
                };

                var jsonPayload = JsonConvert.SerializeObject(requestPayload);
                var content = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

                if (ChatBot.ApiUrl.Contains("localhost") || ChatBot.ApiUrl.Contains("127.0.0.1") || ChatBot.ApiUrl.Contains("10.3.74.59") && ChatBot.ModelName.Contains(":"))
                {
                    if (!string.IsNullOrEmpty(ChatBot.ApiKey))
                    {
                        client.DefaultRequestHeaders.Add("Authorization", $"Bearer {ChatBot.ApiKey}");
                    }

                    try
                    {
                        var response = await client.PostAsync(ChatBot.ApiUrl, content);
                        response.EnsureSuccessStatusCode();

                        using (var stream = await response.Content.ReadAsStreamAsync())
                        using (var reader = new StreamReader(stream))
                        {
                            string line;
                            while ((line = await reader.ReadLineAsync()) != null)
                            {
                                if (!string.IsNullOrWhiteSpace(line))
                                {
                                    var line_trimed = line.Replace("data: ", "");
                                    var responseJson = JObject.Parse(line_trimed);
                                    var contentField = responseJson["message"]?["content"]?.ToString().ToString();

                                    if (!string.IsNullOrEmpty(contentField))
                                    {
                                        loadingBubble.Text += contentField;
                                        await Task.Delay(100);
                                    }
                                }
                            }
                        }

                        ChatBot.ChatHistory.Add(new { role = "assistant", content = loadingBubble.Text });
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"请求AI服务时发生异常: {ex.Message}");
                        loadingBubble.Text = "AI服务异常";
                    }
                }
                else {
                    if (!string.IsNullOrEmpty(ChatBot.ApiKey))
                    {
                        client.DefaultRequestHeaders.Add("Authorization", $"Bearer {ChatBot.ApiKey}");
                    }

                    try
                    {

                        var response = await client.PostAsync(ChatBot.ApiUrl, content);
                        response.EnsureSuccessStatusCode();
                        bool reasoningContentEnded = false;
                        loadingBubble.Text += "<thinking>\n";

                        using (var stream = await response.Content.ReadAsStreamAsync())
                        using (var reader = new StreamReader(stream))
                        {
                            string line;
                            while ((line = await reader.ReadLineAsync()) != null)
                            {
                                if (!string.IsNullOrWhiteSpace(line) && !line.Contains("[DONE]"))
                                {
                                    var line_trimed = line.Replace("data: ", "");
                                    var responseJson = JObject.Parse(line_trimed);
                                    var reasoningField = responseJson["choices"]?[0]?["delta"]?["reasoning_content"]?.ToString();
                                    var contentField = responseJson["choices"]?[0]?["delta"]?["content"]?.ToString();

                                    if (!string.IsNullOrEmpty(reasoningField))
                                    {
                                        loadingBubble.Text += reasoningField;
                                        await Task.Delay(100);
                                        reasoningContentEnded = true;
                                    }

                                    if (!string.IsNullOrEmpty(contentField))
                                    {
                                        if (reasoningContentEnded)
                                        {
                                            loadingBubble.Text += "\n</thinking>\n";
                                            reasoningContentEnded = false;
                                        }
                                        loadingBubble.Text += contentField;
                                        await Task.Delay(100);
                                    }
                                }
                            }
                        }

                        ChatBot.ChatHistory.Add(new { role = "assistant", content = loadingBubble.Text });
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"请求AI服务时发生异常: {ex.Message}");
                        loadingBubble.Text = "AI服务异常";
                    }
                }
                
            }
        }

        private void OnEntryFocused(object sender, FocusEventArgs e)
        {
            if (PromptText.Text == "输入内容")
            {
                PromptText.Text = string.Empty;
            }
        }

        private void OnEntryUnfocused(object sender, FocusEventArgs e)
        {
            if (string.IsNullOrEmpty(PromptText.Text))
            {
                PromptText.Text = "输入内容";
            }
        }

        private async void OnUploadFileClicked(object sender, EventArgs e)
        {
            var result = await FilePicker.PickAsync();
            if (result != null)
            {
                FileLabel.Text = $"已上传文件: {result.FileName}";

                var fileContent = await File.ReadAllTextAsync(result.FullPath);

                var segments = await EmbeddingProcessor.ProcessTextAsync(fileContent);

                var segmentItems = new List<SegmentItem>();
                foreach (var segment in segments)
                {
                    segmentItems.Add(new SegmentItem
                    {
                        FullText = segment.Text,
                        Vector = segment.Vector,
                        DisplayText = segment.Text.Substring(0, Math.Min(10, segment.Text.Length)) + "..."
                    });
                }

                var fileItem = new FileItem
                {
                    FileName = result.FileName,
                    Segments = segmentItems
                };

                var files = FilesCollectionView.ItemsSource as List<FileItem> ?? new List<FileItem>();
                files.Add(fileItem);
                FilesCollectionView.ItemsSource = files;
            }
        }

        // Generate image from text
        private async void OnGenerateImage(object sender, EventArgs e)
        {
            try
            {
                LoadingIndicator.IsRunning = true;
                StatusLabel.IsVisible = false;
                GeneratedImage.Source = null;

                var prompt = PromptEntry.Text;
                if (string.IsNullOrWhiteSpace(prompt))
                {
                    StatusLabel.Text = "请输入图片描述";
                    StatusLabel.IsVisible = true;
                    return;
                }
                var response = await TextToImageClient.GenerateImageAsync(prompt);

                if (!string.IsNullOrEmpty(response.Data?.Url))
                {
                    GeneratedImage.Source = ImageSource.FromUri(new Uri(response.Data.Url));
                    await DisplayAlert("生成成功", $"图片ID: {response.Id}", "确定");
                }
                else
                {
                    StatusLabel.Text = "图片生成失败";
                    StatusLabel.IsVisible = true;
                }
            }
            catch (Exception ex)
            {
                StatusLabel.Text = $"发生错误: {ex.Message}";
                StatusLabel.IsVisible = true;
            }
            finally
            {
                LoadingIndicator.IsRunning = false;
            }
        }

        // Settings
        private void OnApiUrlChanged(object sender, TextChangedEventArgs e)
        {
            ChatBot.ApiUrl = e.NewTextValue;
            SaveSettings();
        }

        private void OnApiKeyChanged(object sender, TextChangedEventArgs e)
        {
            ChatBot.ApiKey = e.NewTextValue;
            SaveSettings();
        }

        private void OnModelNameChanged(object sender, TextChangedEventArgs e)
        {
            ChatBot.ModelName = e.NewTextValue;
            SaveSettings();
        }

        private void SaveSettings()
        {
            var settings = new Dictionary<string, string>
            {
                { "ApiUrl", ChatBot.ApiUrl },
                { "ApiKey", ChatBot.ApiKey },
                { "ModelName", ChatBot.ModelName}
            };

            var json = JsonConvert.SerializeObject(settings);
            File.WriteAllText(SettingsFilePath, json);
        }

        private void LoadSettings()
        {
            if (File.Exists(SettingsFilePath))
            {
                var json = File.ReadAllText(SettingsFilePath);
                var settings = JsonConvert.DeserializeObject<Dictionary<string, string>>(json);

                if (settings.ContainsKey("ApiUrl"))
                {
                    ChatBot.ApiUrl = settings["ApiUrl"];
                    ApiUrlEntry.Text = settings["ApiUrl"];
                }

                if (settings.ContainsKey("ApiKey"))
                {
                    ChatBot.ApiKey = settings["ApiKey"];
                    ApiKeyEntry.Text = settings["ApiKey"];
                }

                if (settings.ContainsKey("ModelName"))
                {
                    ChatBot.ModelName = settings["ModelName"];
                    ModelNameEntry.Text = settings["ModelName"];
                }
            }
        }

        private Label CreateChatBubble(string text, bool isUser)
        {
            return new Label
            {
                Text = text,
                BackgroundColor = isUser ? Colors.LightBlue : Colors.LightGray,
                TextColor = Colors.Black,
                Padding = new Thickness(10),
                Margin = new Thickness(5),
                HorizontalOptions = isUser ? LayoutOptions.End : LayoutOptions.Start,
                VerticalOptions = LayoutOptions.Start
            };
        }

        private async void OnSegmentTapped(object sender, EventArgs e)
        {
            var label = sender as Label;
            var segmentItem = label.BindingContext as SegmentItem;

            if (segmentItem != null)
            {
                await DisplayAlert("Segment Details", $"Text: {segmentItem.FullText}\nVector: {string.Join(", ", segmentItem.Vector)}", "OK");
            }
        }
    }

    public class SegmentItem
    {
        public string FullText { get; set; }
        public float[] Vector { get; set; }
        public string DisplayText { get; set; }
    }

    public class FileItem
    {
        public string FileName { get; set; }
        public List<SegmentItem> Segments { get; set; }
    }
}