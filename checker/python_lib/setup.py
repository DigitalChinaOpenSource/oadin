from setuptools import setup, find_packages

setup(
    name="byzechecker",
    version="1.0.0", 
    author="Your Name",
    author_email="your.email@example.com",
    description="Byze 自动安装和初始化工具",
    long_description=open("README.md", encoding="utf-8").read(),
    long_description_content_type="text/markdown",
    packages=find_packages(),
    install_requires=["requests"], 
    python_requires=">=3.6",
)
