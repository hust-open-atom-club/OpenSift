'''
Check if given git links are broken
'''
from http import HTTPStatus
import asyncio
import httpx

TIMEOUT = 100
GIT_LINK_LIST = "link.txt"
NOT_VALID = "url_not_valid.txt"
NOT_AVAILABLE = "repository_not_available.txt"
OUTPUT = "correct_links.txt"
GITHUB_AVAILABLE = "R55ab"
GITLAB_AVAILABLE = "project-code-holder gl-w-full sm:gl-w-auto"
BITBUCKET_AVAILABLE = "css-1ianfu6"
DEFAULT_AVAILABLE = "clone"

def get_links():
    '''get links'''
    with open(GIT_LINK_LIST, "r", encoding="utf-8") as f:
        links = f.readlines()
        links = [link.strip() for link in links]
        links = [
            link[:-4] if link.endswith(".git") else link
            for link in links
            if len(link) > 5
        ]
    return links

def is_url_valid(url):
    '''check if link valid'''
    if ' ' in url or '\n' in url or '#' in url or '/blob/' in url or '/master/' in url:
        return False
    return True

async def check_github_repository(url):
    '''check if github repo alive'''
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(url, timeout=TIMEOUT)
        except Exception as e:
            print(e)
            return False
        return resp.status_code == HTTPStatus.OK and GITHUB_AVAILABLE in resp.text

async def check_gitlab_repository(url):
    '''check if gitlab repo alive'''
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(url, timeout=TIMEOUT)
        except Exception as e:
            print(e)
            return False
        return resp.status_code == HTTPStatus.OK and GITLAB_AVAILABLE in resp.text

async def check_bitbucket_repository(url):
    '''check if gitlab repo alive'''
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(url, timeout=TIMEOUT)
        except Exception as e:
            print(e)
            return False
        return resp.status_code == HTTPStatus.OK and BITBUCKET_AVAILABLE in resp.text

async def check_other_repository(url):
    '''check if other repo alive'''
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(url, timeout=TIMEOUT)
        except Exception as e:
            print(e)
            return False
        return resp.status_code == HTTPStatus.OK and DEFAULT_AVAILABLE in resp.text.lower()

async def is_repository_alive(url):
    '''check if repository is alive'''
    if "github.com" in url:
        return await check_github_repository(url)
    elif "gitlab.com" in url:
        return await check_gitlab_repository(url)
    elif "bitbucket.org" in url:
        return await check_bitbucket_repository(url)
    else:
        return await check_other_repository(url)

async def main():
    '''entrance'''
    links = get_links()
    invalid_links = []
    moved_links = []
    correct_links = []
    ssh_links = []

    async def process_link(link):
        valid = is_url_valid(link)
        if not valid:
            invalid_links.append(link)
            return

        if "git://" in link or "svn://" in link:
            ssh_links.append(link)
            return

        alive = await is_repository_alive(link)
        if not alive:
            moved_links.append(link)
        else:
            correct_links.append(link)

    tasks = [process_link(link) for link in links]
    await asyncio.gather(*tasks)

    with open(NOT_VALID, "w", encoding="utf-8") as f:
        f.writelines(f'\"{link}.git\"\n' for link in invalid_links)

    with open(NOT_AVAILABLE, "w", encoding="utf-8") as f:
        f.writelines(f'\"{link}.git\"\n' for link in moved_links)

    with open(OUTPUT, "w", encoding="utf-8") as f:
        #* Most ssh links are correct according to current data
        correct_links.extend([link.replace(' ','') for link in ssh_links if link])
        f.writelines(f'\"{link}.git\",\n' for link in correct_links)

if __name__ == "__main__":
    asyncio.run(main())
